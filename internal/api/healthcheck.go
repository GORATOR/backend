package api

import (
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
)

func Healthscheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	model := models.Healthcheck{
		Version:         "1",
		Status:          "OK",
		PostgresVersion: database.GetDatabaseVersion(),
	}
	utils.HttpReturnJson(w, model)
}
