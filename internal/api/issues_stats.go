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
	"gorm.io/gorm"
)

type IssueStatEntry struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type statsParams struct {
	interval    string
	periods     int
	startDate   time.Time
	endDate     time.Time
	sqlInterval string
	timeFormat  string
}

func parseStatsParams(r *http.Request) statsParams {
	interval := "day"
	if intervalParam := utils.GetQueryParam(r, "interval"); intervalParam != "" {
		switch intervalParam {
		case "minute", "hour", "day", "week":
			interval = intervalParam
		}
	}

	periods := 14
	if periodsParam := utils.GetQueryParam(r, "periods"); periodsParam != "" {
		if val, err := parseIntParam(periodsParam); err == nil && val > 0 && val <= 1000 {
			periods = val
		}
	}

	endDate := time.Now()
	var startDate time.Time
	var sqlInterval string
	var timeFormat string

	switch interval {
	case "minute":
		startDate = endDate.Add(-time.Duration(periods-1) * time.Minute)
		startDate = startDate.Truncate(time.Minute)
		sqlInterval = "date_trunc('minute', created_at)"
		timeFormat = "2006-01-02T15:04:00Z"
	case "hour":
		startDate = endDate.Add(-time.Duration(periods-1) * time.Hour)
		startDate = startDate.Truncate(time.Hour)
		sqlInterval = "date_trunc('hour', created_at)"
		timeFormat = "2006-01-02T15:00:00Z"
	case "week":
		startDate = endDate.AddDate(0, 0, -(periods-1)*7)
		weekday := int(startDate.Weekday())
		if weekday == 0 {
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

	return statsParams{
		interval:    interval,
		periods:     periods,
		startDate:   startDate,
		endDate:     endDate,
		sqlInterval: sqlInterval,
		timeFormat:  timeFormat,
	}
}

func buildStatsBaseQuery(db *gorm.DB, params statsParams) *gorm.DB {
	return db.Table("envelope_event_commons").
		Select(fmt.Sprintf("%s as date, COUNT(*) as count", params.sqlInterval)).
		Where("deleted_at IS NULL").
		Where("created_at >= ?", params.startDate).
		Where("created_at <= ?", params.endDate)
}

func scanAndFillStats(query *gorm.DB, params statsParams) ([]IssueStatEntry, error) {
	var stats []IssueStatEntry

	result := query.
		Group(params.sqlInterval).
		Order("date ASC").
		Scan(&stats)

	if result.Error != nil {
		return nil, result.Error
	}

	statsMap := make(map[string]int64)
	for _, stat := range stats {
		var dateKey string

		if t, err := time.Parse(time.RFC3339, stat.Date); err == nil {
			if params.interval == "week" {
				weekday := int(t.Weekday())
				if weekday == 0 {
					weekday = 7
				}
				daysToMonday := weekday - 1
				weekStart := t.AddDate(0, 0, -daysToMonday)
				dateKey = weekStart.Format(params.timeFormat)
			} else {
				dateKey = t.Format(params.timeFormat)
			}
			statsMap[dateKey] = stat.Count
		} else if t, err := time.Parse("2006-01-02", stat.Date); err == nil {
			dateKey = t.Format(params.timeFormat)
			statsMap[dateKey] = stat.Count
		} else {
			statsMap[stat.Date] = stat.Count
		}
	}

	var fullStats []IssueStatEntry
	currentTime := params.startDate

	for i := 0; i < params.periods; i++ {
		var dateKey string
		switch params.interval {
		case "minute":
			dateKey = currentTime.Format(params.timeFormat)
			currentTime = currentTime.Add(time.Minute)
		case "hour":
			dateKey = currentTime.Format(params.timeFormat)
			currentTime = currentTime.Add(time.Hour)
		case "week":
			weekday := int(currentTime.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			daysToMonday := weekday - 1
			weekStart := currentTime.AddDate(0, 0, -daysToMonday)
			weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
			dateKey = weekStart.Format(params.timeFormat)
			currentTime = currentTime.AddDate(0, 0, 7)
		default:
			dateKey = currentTime.Format(params.timeFormat)
			currentTime = currentTime.AddDate(0, 0, 1)
		}

		count := statsMap[dateKey]
		fullStats = append(fullStats, IssueStatEntry{
			Date:  dateKey,
			Count: count,
		})
	}

	return fullStats, nil
}

func IssuesStats(w http.ResponseWriter, r *http.Request) {
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

	db := database.GetDatabaseConnection()
	params := parseStatsParams(r)
	queryBuilder := buildStatsBaseQuery(db, params)

	eventTypeFilter := utils.GetQueryParam(r, "eventType")

	switch eventTypeFilter {
	case EventTypeException:
		queryBuilder = queryBuilder.Where("exception_type IS NOT NULL AND exception_type != ''")
	case EventTypeMessage:
		queryBuilder = queryBuilder.Where("exception_type IS NULL OR exception_type = ''")
	}

	if len(projectIds) > 0 {
		queryBuilder = queryBuilder.Where("project_id IN ?", projectIds)
	}

	fullStats, err := scanAndFillStats(queryBuilder, params)
	if err != nil {
		fmt.Printf("Error querying issue stats: %v\n", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.HttpReturnJson(w, fullStats)
}

func parseIntParam(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}
