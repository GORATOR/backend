package crud

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
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

		forbiddenActionStr := fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead)

		userId, ok := before(w, r, m, nil)
		if !ok {
			http.Error(w, forbiddenActionStr, http.StatusForbidden)
			return
		}

		if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, m.GetName()) {
			http.Error(w, forbiddenActionStr, http.StatusForbidden)
			return
		}

		query := database.GetDatabaseConnection().Model(&m)
		m.ParseQueryString(r.URL.RawPath, query, r)
		groupBy := utils.GetQueryParam(r, "groupBy")
		if groupBy != "" {
			countEntitiesGroupedResult(w, groupBy, query, m)
		} else {
			countEntitiesResult(w, query, m.GetName())
		}
	}
}

func onCountError(modelName string, w http.ResponseWriter, countResult *gorm.DB) {
	fmt.Printf("CountEntities error for \"'%s\" entity ", modelName)
	fmt.Print(countResult.Error)
	http.Error(w, "", http.StatusBadRequest)
}

func countEntitiesGroupedResult(w http.ResponseWriter, groupBy string, query *gorm.DB, m models.ReadableModel) {
	var result []models.ModelGroupedCountRecord
	if !m.IsAllowedGroupField(groupBy) {
		fmt.Printf("countEntitiesGroupedResult: using disallowed field %s", groupBy)
		http.Error(w, fmt.Sprintf("Group by %s field is not allowed", groupBy), http.StatusBadRequest)
		return
	}
	selectStr := fmt.Sprintf("count(*) AS count, %s AS field", groupBy)
	countResult := query.Select(selectStr).Scan(&result)

	if countResult.Error != nil {
		onCountError(m.GetName(), w, countResult)
		return
	}

	response := models.ModelGroupedCountResponse{
		Entity:  m.GetName(),
		GroupBy: groupBy,
		Data:    result,
	}
	utils.HttpReturnJson(w, response)
}

func countEntitiesResult(w http.ResponseWriter, query *gorm.DB, modelName string) {
	var count int64
	countResult := query.Count(&count)

	if countResult.Error != nil {
		onCountError(modelName, w, countResult)
		return
	}

	response := models.ModelCountResponse{
		Entity: modelName,
		Count:  count,
	}
	utils.HttpReturnJson(w, response)
}
