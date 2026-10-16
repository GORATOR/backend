package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

const maxTracesLimit = 100

type traceGroup struct {
	TraceId string
	LastID  uint
}

// applyTraceFilters is the single source of truth for /traces filters.
// Both /traces and /traces/count must build their queries from it,
// otherwise the count stops matching the paginated list.
func applyTraceFilters(r *http.Request) *gorm.DB {
	// Transactions without a trace_id are not traces: they must not
	// occupy page slots or be counted.
	db := database.GetDatabaseConnection().
		Model(&models.Transaction{}).
		Where("trace_id IS NOT NULL AND trace_id <> ''")

	if projectId := utils.GetQueryParam(r, "projectId"); projectId != "" {
		db = db.Where("project_id = ?", projectId)
	}
	if traceId := utils.GetQueryParam(r, "traceId"); traceId != "" {
		db = db.Where("trace_id = ?", traceId)
	}
	if from := utils.GetQueryParam(r, "from"); from != "" {
		db = db.Where("start_timestamp >= ?", from)
	}
	if to := utils.GetQueryParam(r, "to"); to != "" {
		db = db.Where("start_timestamp <= ?", to)
	}
	return db
}

func Traces(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}
	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.TransactionModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	limit := 10
	offset := 0
	if limitStr := utils.GetQueryParam(r, "limit"); limitStr != "" {
		// only positive values are meaningful; negative ones used to panic
		// on the in-memory slice
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}
	if offsetStr := utils.GetQueryParam(r, "offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val > 0 {
			offset = val
		}
	}
	if limit > maxTracesLimit {
		limit = maxTracesLimit
	}

	var groups []traceGroup
	if result := applyTraceFilters(r).
		Select("trace_id, MAX(id) as last_id").
		Group("trace_id").
		Order("last_id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&groups); result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	traceIds := make([]string, 0, len(groups))
	for _, g := range groups {
		traceIds = append(traceIds, g.TraceId)
	}

	txs := make([]models.Transaction, 0)
	if len(traceIds) > 0 {
		// same filters as the grouping query: without the project filter
		// cross-project traces would pull in transactions of other projects
		if result := applyTraceFilters(r).
			Where("trace_id IN ?", traceIds).
			Preload("Project").
			Order("id").
			Find(&txs); result.Error != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	utils.HttpReturnJson(w, txs)
}

func TracesCount(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}
	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.TransactionModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	var count int64
	if result := applyTraceFilters(r).Distinct("trace_id").Count(&count); result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.HttpReturnJson(w, models.ModelCountResponse{
		Entity: models.TransactionModelName,
		Count:  count,
	})
}

func Trace(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}
	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.TransactionModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	traceId := r.PathValue("trace_id")
	if traceId == "" {
		utils.HttpReturnBadRequest(w)
		return
	}

	var txs []models.Transaction
	database.GetDatabaseConnection().
		Where("trace_id = ?", traceId).
		Preload("Spans").
		Preload("Project").
		Find(&txs)

	utils.HttpReturnJson(w, map[string]interface{}{
		"trace_id":     traceId,
		"transactions": txs,
	})
}
