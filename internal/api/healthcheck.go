package api

import (
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
)

func Healthscheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	pgVersion := database.GetDatabaseVersion()
	var status string
	if pgVersion == "" {
		status = "Database connection isn't active"
	} else {
		status = "OK"
	}
	model := models.Healthcheck{
		Version:         "1",
		Status:          status,
		PostgresVersion: pgVersion,
	}
	utils.HttpReturnJson(w, model)
}
