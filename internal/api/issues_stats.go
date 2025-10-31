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

type IssueStatEntry struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

func IssuesStats(w http.ResponseWriter, r *http.Request) {
	// Check authorization
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, models.EnvelopeEventCommonModelName) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return
	}

	// Parse interval parameter: minute, hour, day, week (default: day)
	interval := "day"
	if intervalParam := utils.GetQueryParam(r, "interval"); intervalParam != "" {
		switch intervalParam {
		case "minute", "hour", "day", "week":
			interval = intervalParam
		}
	}

	// Parse periods parameter (number of intervals to return)
	periods := 14
	if periodsParam := utils.GetQueryParam(r, "periods"); periodsParam != "" {
		if val, err := parseIntParam(periodsParam); err == nil && val > 0 && val <= 1000 {
			periods = val
		}
	}

	// Parse projectIds parameter
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

	db := database.GetDatabaseConnection()

	// Calculate date range based on interval and periods
	endDate := time.Now()
	var startDate time.Time
	var sqlInterval string
	var timeFormat string

	switch interval {
	case "minute":
		startDate = endDate.Add(-time.Duration(periods-1) * time.Minute)
		// Truncate to minute
		startDate = startDate.Truncate(time.Minute)
		sqlInterval = "date_trunc('minute', created_at)"
		timeFormat = "2006-01-02T15:04:00Z"
	case "hour":
		startDate = endDate.Add(-time.Duration(periods-1) * time.Hour)
		// Truncate to hour
		startDate = startDate.Truncate(time.Hour)
		sqlInterval = "date_trunc('hour', created_at)"
		timeFormat = "2006-01-02T15:00:00Z"
	case "week":
		startDate = endDate.AddDate(0, 0, -(periods-1)*7)
		// Truncate to start of week (Monday)
		weekday := int(startDate.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		daysToMonday := weekday - 1
		startDate = startDate.AddDate(0, 0, -daysToMonday)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		sqlInterval = "date_trunc('week', created_at)"
		timeFormat = "2006-01-02"
	default: // day
		startDate = endDate.AddDate(0, 0, -periods+1)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		sqlInterval = "created_at::date"
		timeFormat = "2006-01-02"
	}

	var stats []IssueStatEntry

	// Build query with optional project filter
	queryBuilder := db.Table("envelope_event_commons").
		Select(fmt.Sprintf("%s as date, COUNT(*) as count", sqlInterval)).
		Where("deleted_at IS NULL").
		Where("created_at >= ?", startDate).
		Where("created_at <= ?", endDate)

	// Apply project filter if projectIds are provided
	if len(projectIds) > 0 {
		queryBuilder = queryBuilder.Where("project_id IN ?", projectIds)
	}

	result := queryBuilder.
		Group(sqlInterval).
		Order("date ASC").
		Scan(&stats)

	if result.Error != nil {
		fmt.Printf("Error querying issue stats: %v\n", result.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Fill in missing intervals with zero counts
	statsMap := make(map[string]int64)
	for _, stat := range stats {
		// Parse timestamp and convert to appropriate format
		var dateKey string

		// Try parsing as RFC3339 (returned by date_trunc)
		if t, err := time.Parse(time.RFC3339, stat.Date); err == nil {
			if interval == "week" {
				// For weeks, ensure we use the Monday start
				weekday := int(t.Weekday())
				if weekday == 0 {
					weekday = 7
				}
				daysToMonday := weekday - 1
				weekStart := t.AddDate(0, 0, -daysToMonday)
				dateKey = weekStart.Format(timeFormat)
			} else {
				dateKey = t.Format(timeFormat)
			}
			statsMap[dateKey] = stat.Count
		} else if t, err := time.Parse("2006-01-02", stat.Date); err == nil {
			dateKey = t.Format(timeFormat)
			statsMap[dateKey] = stat.Count
		} else {
			// Fallback: use as-is
			statsMap[stat.Date] = stat.Count
		}
	}

	// Generate full stats with all intervals
	var fullStats []IssueStatEntry
	currentTime := startDate

	for i := 0; i < periods; i++ {
		var dateKey string
		switch interval {
		case "minute":
			dateKey = currentTime.Format(timeFormat)
			currentTime = currentTime.Add(time.Minute)
		case "hour":
			dateKey = currentTime.Format(timeFormat)
			currentTime = currentTime.Add(time.Hour)
		case "week":
			// Ensure we're at the start of the week
			weekday := int(currentTime.Weekday())
			if weekday == 0 { // Sunday
				weekday = 7
			}
			daysToMonday := weekday - 1
			weekStart := currentTime.AddDate(0, 0, -daysToMonday)
			weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
			dateKey = weekStart.Format(timeFormat)
			currentTime = currentTime.AddDate(0, 0, 7)
		default: // day
			dateKey = currentTime.Format(timeFormat)
			currentTime = currentTime.AddDate(0, 0, 1)
		}

		count := statsMap[dateKey]
		fullStats = append(fullStats, IssueStatEntry{
			Date:  dateKey,
			Count: count,
		})
	}

	utils.HttpReturnJson(w, fullStats)
}

func parseIntParam(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}
