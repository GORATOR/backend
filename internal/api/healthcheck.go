package api

import (
	"net/http"

	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
)

func Healscheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	model := models.Healthcheck{Version: "1", Status: "OK", PostgresVersion: "16"}
	utils.HttpReturnJson(w, model)
}
