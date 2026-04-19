package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

func buildGroupQuery(db *gorm.DB, envelope *models.EnvelopeEventCommon) *gorm.DB {
	query := db.Where("deleted_at IS NULL")

	if envelope.ExceptionType != "" {
		query = query.Where("exception_type = ?", envelope.ExceptionType)
		if envelope.StacktraceHash != "" {
			query = query.Where("stacktrace_hash = ?", envelope.StacktraceHash)
		} else {
			query = query.Where("exception_value = ?", envelope.ExceptionValue)
		}
	} else {
		query = query.Where("message = ? AND level = ?", envelope.Message, envelope.Level)
	}

	return query
}

func IssueEvents(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.EnvelopeEventCommonModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	db := database.GetDatabaseConnection()

	var envelope models.EnvelopeEventCommon
	if result := db.Where("id = ?", id).First(&envelope); result.Error != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	limit := 10
	offset := 0
	sortOrder := "DESC"

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

	if sortOrderParam := utils.GetQueryParam(r, "sortOrder"); sortOrderParam != "" {
		if sortOrderParam == "ASC" || sortOrderParam == "DESC" {
			sortOrder = sortOrderParam
		}
	}

	query := buildGroupQuery(db.Table("envelope_event_commons"), &envelope)

	if createdAtFromParam := utils.GetQueryParam(r, "createdAtFrom"); createdAtFromParam != "" {
		if t, err := time.Parse(time.RFC3339, createdAtFromParam); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}

	if createdAtToParam := utils.GetQueryParam(r, "createdAtTo"); createdAtToParam != "" {
		if t, err := time.Parse(time.RFC3339, createdAtToParam); err == nil {
			query = query.Where("created_at <= ?", t)
		}
	}

	query = query.Order("created_at " + sortOrder).
		Limit(limit).
		Offset(offset)

	var records []models.EnvelopeEventCommon
	if result := query.
		Preload("EventCommonSdk").
		Preload("EnvelopeEventExtras").
		Preload("Project").
		Find(&records); result.Error != nil {
		fmt.Printf("Error querying issue events: %v\n", result.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.HttpReturnJson(w, records)
}

func IssueEventsStats(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.EnvelopeEventCommonModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	db := database.GetDatabaseConnection()

	var envelope models.EnvelopeEventCommon
	if result := db.Where("id = ?", id).First(&envelope); result.Error != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	params := parseStatsParams(r)

	// Find the earliest event in this group to cover the full time range
	var firstCreatedAt time.Time
	firstQuery := buildGroupQuery(db.Table("envelope_event_commons"), &envelope)
	firstQuery.Select("MIN(created_at)").Row().Scan(&firstCreatedAt)

	if !firstCreatedAt.IsZero() && firstCreatedAt.Before(params.startDate) {
		params.startDate = firstCreatedAt
		switch params.interval {
		case "minute":
			params.startDate = params.startDate.Truncate(time.Minute)
			params.periods = int(params.endDate.Sub(params.startDate).Minutes()) + 1
		case "hour":
			params.startDate = params.startDate.Truncate(time.Hour)
			params.periods = int(params.endDate.Sub(params.startDate).Hours()) + 1
		case "week":
			weekday := int(params.startDate.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			params.startDate = params.startDate.AddDate(0, 0, -(weekday - 1))
			params.startDate = time.Date(params.startDate.Year(), params.startDate.Month(), params.startDate.Day(), 0, 0, 0, 0, params.startDate.Location())
			params.periods = int(params.endDate.Sub(params.startDate).Hours()/(24*7)) + 1
		default: // day
			params.startDate = time.Date(params.startDate.Year(), params.startDate.Month(), params.startDate.Day(), 0, 0, 0, 0, params.startDate.Location())
			params.periods = int(params.endDate.Sub(params.startDate).Hours()/24) + 1
		}
	}

	queryBuilder := buildStatsBaseQuery(db, params)
	queryBuilder = buildGroupQuery(queryBuilder, &envelope)

	fullStats, err := scanAndFillStats(queryBuilder, params)
	if err != nil {
		fmt.Printf("Error querying issue events stats: %v\n", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.HttpReturnJson(w, fullStats)
}

func IssueEventsCount(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.EnvelopeEventCommonModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	db := database.GetDatabaseConnection()

	var envelope models.EnvelopeEventCommon
	if result := db.Where("id = ?", id).First(&envelope); result.Error != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	query := buildGroupQuery(db.Table("envelope_event_commons"), &envelope)

	var count int64
	if result := query.Count(&count); result.Error != nil {
		fmt.Printf("Error counting issue events: %v\n", result.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := models.ModelCountResponse{
		Count:  count,
		Entity: "envelope",
	}

	utils.HttpReturnJson(w, response)
}
