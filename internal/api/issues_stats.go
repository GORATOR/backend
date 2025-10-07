package api

import (
	"fmt"
	"net/http"
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

	// Default to last 14 days
	days := 14
	if daysParam := utils.GetQueryParam(r, "days"); daysParam != "" {
		if val, err := parseIntParam(daysParam); err == nil && val > 0 && val <= 90 {
			days = val
		}
	}

	db := database.GetDatabaseConnection()

	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	var stats []IssueStatEntry

	// Query to get event count per day
	result := db.Raw(`
		SELECT
			created_at::date as date,
			COUNT(*) as count
		FROM envelope_event_commons
		WHERE deleted_at IS NULL
			AND created_at::date >= ?::date
			AND created_at::date <= ?::date
		GROUP BY created_at::date
		ORDER BY date ASC
	`, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Scan(&stats)

	if result.Error != nil {
		fmt.Printf("Error querying issue stats: %v\n", result.Error)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Fill in missing dates with zero counts
	statsMap := make(map[string]int64)
	for _, stat := range stats {
		// Parse timestamp and convert to date string
		if t, err := time.Parse(time.RFC3339, stat.Date); err == nil {
			dateKey := t.Format("2006-01-02")
			statsMap[dateKey] = stat.Count
		} else {
			// Fallback: use as-is if it's already in date format
			statsMap[stat.Date] = stat.Count
		}
	}

	var fullStats []IssueStatEntry
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		count := statsMap[date]
		fullStats = append(fullStats, IssueStatEntry{
			Date:  date,
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
