package crud

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

const (
	defaultLimit  = 10
	defaultOffset = 0
)

func Read(m models.ReadableModel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		userId, ok := before(w, r, m.GetName(), &id)
		if !ok {
			return
		}

		if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, m.GetName()) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}

		data, err := m.ReadById(database.GetDatabaseConnection(), id)
		if err != nil {
			http.Error(w, "DB error", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(data)
		w.WriteHeader(http.StatusOK)

	}
}

func ReadEntities(m models.ReadableModel, ignoreActive bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := tryBuildReadQuery(w, r, m, ignoreActive)
		if err != nil {
			return
		}
		m.ParseQueryString("ReadEntities", query, r)
		entities, err := m.FindAll(query)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.HttpReturnJson(w, entities)
	}
}

func CountEntities(m models.ReadableModel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var count int64

		forbiddenActionStr := fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead)

		userId, ok := before(w, r, m.GetName(), nil)
		if !ok {
			http.Error(w, forbiddenActionStr, http.StatusForbidden)
			return
		}

		if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, m.GetName()) {
			http.Error(w, forbiddenActionStr, http.StatusForbidden)
			return
		}

		query := database.GetDatabaseConnection().Model(&m)
		countResult := query.Count(&count)

		if countResult.Error != nil {
			fmt.Printf("CountEntities error for %s entity", m.GetName())
			fmt.Print(countResult.Error)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		response := models.ModelCountResponse{
			Entity: m.GetName(),
			Count:  count,
		}
		utils.HttpReturnJson(w, response)
	}
}
