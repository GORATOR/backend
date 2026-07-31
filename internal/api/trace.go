package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

type traceGroup struct {
	TraceId string
	LastID  uint
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
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}
	if offsetStr := utils.GetQueryParam(r, "offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}

	db := database.GetDatabaseConnection().Model(&models.Transaction{})

	if projectId := utils.GetQueryParam(r, "projectId"); projectId != "" {
		db.Where("project_id = ?", projectId)
	}
	if traceId := utils.GetQueryParam(r, "traceId"); traceId != "" {
		db.Where("trace_id = ?", traceId)
	}
	if from := utils.GetQueryParam(r, "from"); from != "" {
		db.Where("start_timestamp >= ?", from)
	}
	if to := utils.GetQueryParam(r, "to"); to != "" {
		db.Where("start_timestamp <= ?", to)
	}

	var groups []traceGroup
	if result := db.Select("trace_id, MAX(id) as last_id").
		Group("trace_id").
		Order("last_id DESC").
		Scan(&groups); result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if offset < len(groups) {
		groups = groups[offset:]
	} else {
		groups = nil
	}
	if limit < len(groups) {
		groups = groups[:limit]
	}

	traceIds := make([]string, 0, len(groups))
	for _, g := range groups {
		traceIds = append(traceIds, g.TraceId)
	}

	txs := make([]models.Transaction, 0)
	if len(traceIds) > 0 {
		database.GetDatabaseConnection().
			Where("trace_id IN ?", traceIds).
			Preload("Project").
			Find(&txs)
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

	db := database.GetDatabaseConnection().Model(&models.Transaction{})

	if projectId := utils.GetQueryParam(r, "projectId"); projectId != "" {
		db.Where("project_id = ?", projectId)
	}

	var count int64
	db.Distinct("trace_id").Count(&count)

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
