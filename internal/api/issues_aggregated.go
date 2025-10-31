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

type AggregatedIssue struct {
	Envelope models.EnvelopeEventCommon `json:"envelope"`
	Count    int64                      `json:"count"`
}

func IssuesAggregated(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.EnvelopeEventCommonModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	limit := 10
	offset := 0
	sortBy := "first_id"
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

	if sortByParam := utils.GetQueryParam(r, "sortBy"); sortByParam != "" {
		if sortByParam == "count" {
			sortBy = "count"
		}
	}

	if sortOrderParam := utils.GetQueryParam(r, "sortOrder"); sortOrderParam != "" {
		if sortOrderParam == "ASC" || sortOrderParam == "DESC" {
			sortOrder = sortOrderParam
		}
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

	type IssueGroup struct {
		ExceptionType  string
		ExceptionValue string
		Count          int64
		FirstID        uint
	}

	var groups []IssueGroup

	query := db.Table("envelope_event_commons").
		Select(`
			exception_type,
			exception_value,
			COUNT(*) as count,
			MIN(id) as first_id
		`).
		Where("deleted_at IS NULL").
		Where("exception_type IS NOT NULL").
		Where("exception_type != ''")

	if createdAtFrom != nil {
		query = query.Where("created_at >= ?", *createdAtFrom)
	}

	if len(projectIds) > 0 {
		query = query.Where("project_id IN ?", projectIds)
	}

	result := query.
		Group("exception_type, exception_value").
		Order(sortBy + " " + sortOrder).
		Limit(limit).
		Offset(offset).
		Scan(&groups)

	if result.Error != nil {
		fmt.Printf("Error querying aggregated issues: %v\n", result.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var issues []AggregatedIssue
	for _, group := range groups {
		var envelope models.EnvelopeEventCommon
		err := db.Where("id = ?", group.FirstID).
			Preload("EventCommonSdk").
			Preload("EnvelopeEventExtras").
			Preload("Project").
			First(&envelope).Error

		if err != nil {
			fmt.Printf("Error loading envelope %d: %v\n", group.FirstID, err)
			continue
		}

		issues = append(issues, AggregatedIssue{
			Envelope: envelope,
			Count:    group.Count,
		})
	}

	utils.HttpReturnJson(w, issues)
}
