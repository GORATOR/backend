package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

func IssuesAggregatedCount(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.EnvelopeEventCommonModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	var projectIds []uint
	if projectIdsParam := utils.GetQueryParam(r, "projectIds"); projectIdsParam != "" {
		projectIdsStr := strings.Split(projectIdsParam, ",")
		for _, idStr := range projectIdsStr {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				projectIds = append(projectIds, uint(id))
			}
		}
	}

	var createdAtFrom *time.Time
	if createdAtFromParam := utils.GetQueryParam(r, "createdAtFrom"); createdAtFromParam != "" {
		if t, err := time.Parse(time.RFC3339, createdAtFromParam); err == nil {
			createdAtFrom = &t
		}
	}

	eventTypeFilter := utils.GetQueryParam(r, "eventType")

	db := database.GetDatabaseConnection()

	var exceptionCount int64
	var messageCount int64

	exceptionQuery := db.Table("envelope_event_commons").
		Select("COUNT(DISTINCT (exception_type, COALESCE(NULLIF(stacktrace_hash, ''), exception_value)))").
		Where("deleted_at IS NULL").
		Where("exception_type IS NOT NULL AND exception_type != ''")

	if createdAtFrom != nil {
		exceptionQuery = exceptionQuery.Where("created_at >= ?", *createdAtFrom)
	}
	if len(projectIds) > 0 {
		exceptionQuery = exceptionQuery.Where("project_id IN ?", projectIds)
	}

	if eventTypeFilter != EventTypeMessage {
		if result := exceptionQuery.Scan(&exceptionCount); result.Error != nil {
			fmt.Printf("Error counting exception groups: %v\n", result.Error)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	if eventTypeFilter != EventTypeException {
		messageQuery := db.Table("envelope_event_commons").
			Select("COUNT(DISTINCT message)").
			Where("deleted_at IS NULL").
			Where("(exception_type IS NULL OR exception_type = '') AND message IS NOT NULL AND message != ''")

		if createdAtFrom != nil {
			messageQuery = messageQuery.Where("created_at >= ?", *createdAtFrom)
		}
		if len(projectIds) > 0 {
			messageQuery = messageQuery.Where("project_id IN ?", projectIds)
		}

		if result := messageQuery.Scan(&messageCount); result.Error != nil {
			fmt.Printf("Error counting message groups: %v\n", result.Error)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	count := exceptionCount + messageCount

	response := models.ModelCountResponse{
		Count:  count,
		Entity: "envelope",
	}

	utils.HttpReturnJson(w, response)
}
