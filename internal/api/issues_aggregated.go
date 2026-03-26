package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

type AggregatedIssue struct {
	Envelope  models.EnvelopeEventCommon `json:"envelope"`
	Count     int64                      `json:"count"`
	EventType string                     `json:"event_type"`
}

type issueGroup struct {
	ExceptionType  string
	ExceptionValue string
	Message        string
	Level          string
	Count          int64
	LastID         uint
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
	sortBy := "last_id"
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

	eventTypeFilter := utils.GetQueryParam(r, "eventType")

	db := database.GetDatabaseConnection()

	var exceptionGroups []issueGroup
	var messageGroups []issueGroup

	// --- exceptions ---
	exceptionQuery := db.Table("envelope_event_commons").
		Select(`exception_type, MIN(exception_value) as exception_value, '' as message, '' as level, COUNT(*) as count, MAX(id) as last_id`).
		Where("deleted_at IS NULL").
		Where("exception_type IS NOT NULL AND exception_type != ''")

	if createdAtFrom != nil {
		exceptionQuery = exceptionQuery.Where("created_at >= ?", *createdAtFrom)
	}
	if len(projectIds) > 0 {
		exceptionQuery = exceptionQuery.Where("project_id IN ?", projectIds)
	}

	if eventTypeFilter != EventTypeMessage {
		groupExpr := "exception_type, COALESCE(NULLIF(stacktrace_hash, ''), exception_value)"
		if result := exceptionQuery.Group(groupExpr).Scan(&exceptionGroups); result.Error != nil {
			fmt.Printf("Error querying exception groups: %v\n", result.Error)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	// --- messages ---
	if eventTypeFilter != EventTypeException {
		messageQuery := db.Table("envelope_event_commons").
			Select(`'' as exception_type, '' as exception_value, message, level, COUNT(*) as count, MAX(id) as last_id`).
			Where("deleted_at IS NULL").
			Where("(exception_type IS NULL OR exception_type = '') AND message IS NOT NULL AND message != ''")

		if createdAtFrom != nil {
			messageQuery = messageQuery.Where("created_at >= ?", *createdAtFrom)
		}
		if len(projectIds) > 0 {
			messageQuery = messageQuery.Where("project_id IN ?", projectIds)
		}

		if result := messageQuery.Group("message, level").Scan(&messageGroups); result.Error != nil {
			fmt.Printf("Error querying message groups: %v\n", result.Error)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	allGroups := append(exceptionGroups, messageGroups...)

	sort.Slice(allGroups, func(i, j int) bool {
		var less bool
		if sortBy == "count" {
			less = allGroups[i].Count < allGroups[j].Count
		} else {
			less = allGroups[i].LastID < allGroups[j].LastID
		}
		if sortOrder == "DESC" {
			return !less
		}
		return less
	})

	if offset < len(allGroups) {
		allGroups = allGroups[offset:]
	} else {
		allGroups = nil
	}
	if limit < len(allGroups) {
		allGroups = allGroups[:limit]
	}

	issues := make([]AggregatedIssue, 0)
	for _, group := range allGroups {
		var envelope models.EnvelopeEventCommon
		err := db.Where("id = ?", group.LastID).
			Preload("EventCommonSdk").
			Preload("EnvelopeEventExtras").
			Preload("Project").
			First(&envelope).Error

		if err != nil {
			fmt.Printf("Error loading envelope %d: %v\n", group.LastID, err)
			continue
		}

		eventType := EventTypeException
		if group.ExceptionType == "" {
			eventType = EventTypeMessage
		}

		issues = append(issues, AggregatedIssue{
			Envelope:  envelope,
			Count:     group.Count,
			EventType: eventType,
		})
	}

	utils.HttpReturnJson(w, issues)
}
