package api

import (
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

func IssuesAggregatedCount(w http.ResponseWriter, r *http.Request) {
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

	db := database.GetDatabaseConnection()

	var count int64

	// Count unique exception types and values
	result := db.Table("envelope_event_commons").
		Select("COUNT(DISTINCT (exception_type, exception_value))").
		Where("deleted_at IS NULL").
		Where("exception_type IS NOT NULL").
		Where("exception_type != ''").
		Count(&count)

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
