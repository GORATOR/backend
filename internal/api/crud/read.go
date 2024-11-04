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

func Read(m models.Model) http.HandlerFunc {
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

		data, err := database.GetRecord(id, m)
		if err != nil {
			http.Error(w, "DB error", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(data)
		w.WriteHeader(http.StatusOK)

	}
}

func tryGetRecords(selectFields []string, query *gorm.DB, entities *[]models.Model) error {
	if selectFields != nil {
		query.Select(selectFields)
	}
	result := query.Find(&entities)
	if result.Error != nil {
		fmt.Print("tryGetRecords query.Find error ", result.Error)
		return result.Error
	}
	return nil
}

func ReadUsers(w http.ResponseWriter, r *http.Request) {

	query, err := tryBuildReadQuery(w, r, &models.User{})
	if err != nil {
		return
	}
	var entities []models.Model
	parseUsersQuery(query, r)
	err = tryGetRecords(models.UserSelectFields, query, &entities)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	utils.HttpReturnJson(w, entities)
}

func ReadEntities(m models.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := tryBuildReadQuery(w, r, m)
		if err != nil {
			return
		}
		var entities []models.Model
		parseNameQueryParam(query, r)
		err = tryGetRecords(nil, query, &entities)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.HttpReturnJson(w, entities)
	}
}

func CountEntities(m models.Model) http.HandlerFunc {
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

func parseUsersQuery(query *gorm.DB, r *http.Request) {
	username := utils.GetQueryParam(r, "username")
	if username != "" {
		query.Where("username like ?", username+"%")
	}
}

func parseNameQueryParam(query *gorm.DB, r *http.Request) {
	name := utils.GetQueryParam(r, "name")
	if name != "" {
		query.Where("name like ?", name+"%")
	}
}
