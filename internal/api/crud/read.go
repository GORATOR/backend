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
		userId, ok := before(w, r, m, &id)
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
	}
}

func readError(w http.ResponseWriter, modelName string, err error) {
	fmt.Printf("CountEntities error for \"'%s\" entity ", modelName)
	fmt.Print(err)
	w.WriteHeader(http.StatusBadRequest)
}

func ReadEntities(m models.ReadableModel, ignoreActive bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := tryBuildReadQuery(w, r, m, ignoreActive)
		if err != nil {
			return
		}
		m.ParseQueryString("ReadEntities", query, r)
		entities, err := m.FindAll(query, utils.GetQueryParam(r, "groupBy"))
		if err != nil {
			readError(w, m.GetName(), err)
			return
		}
		utils.HttpReturnJson(w, entities)
	}
}

func CountEntities(m models.ReadableModel, ignoreActive bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := tryBuildReadQuery(w, r, m, ignoreActive)
		if err != nil {
			return
		}
		m.ParseQueryString(r.URL.RawPath, query, r)
		result, err := m.Count(query, utils.GetQueryParam(r, "groupBy"))
		if err != nil {
			readError(w, m.GetName(), err)
			return
		}
		utils.HttpReturnJson(w, result)
	}
}
