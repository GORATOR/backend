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

	db := database.GetDatabaseConnection()

	var count int64

	query := db.Table("envelope_event_commons").
		Select("COUNT(DISTINCT (exception_type, exception_value))").
		Where("deleted_at IS NULL").
		Where("exception_type IS NOT NULL").
		Where("exception_type != ''")

	if createdAtFrom != nil {
		query = query.Where("created_at >= ?", *createdAtFrom)
	}

	if len(projectIds) > 0 {
		query = query.Where("project_id IN ?", projectIds)
	}

	result := query.Count(&count)

	if result.Error != nil {
		fmt.Printf("Error counting aggregated issues: %v\n", result.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := models.ModelCountResponse{
		Count:  count,
		Entity: "envelope",
	}

	utils.HttpReturnJson(w, response)
}
