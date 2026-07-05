package api

import (
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

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

	var txs []models.Transaction
	db.Preload("Project").Find(&txs)

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
	db.Count(&count)

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
