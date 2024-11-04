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

func Read[V models.Model](entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		userId, ok := before(w, r, entity, &id)
		if !ok {
			return
		}

		if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, entity) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}

		data, err := database.GetRecord[V](id)
		if err != nil {
			http.Error(w, "DB error", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(data)
		w.WriteHeader(http.StatusOK)

	}
}

func tryGetRecords[V models.Model](selectFields []string, query *gorm.DB, entities *[]V) error {
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

	query, err := tryBuildReadQuery[models.User](w, r, models.UserModelName)
	if err != nil {
		return
	}
	var entities []models.User
	parseUsersQuery(query, r)
	err = tryGetRecords(models.UserSelectFields, query, &entities)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	utils.HttpReturnJson(w, entities)
}

func ReadEntities[V models.Model](entityName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := tryBuildReadQuery[V](w, r, entityName)
		if err != nil {
			return
		}
		var entities []V
		parseNameQueryParam(query, r)
		err = tryGetRecords(nil, query, &entities)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		utils.HttpReturnJson(w, entities)
	}
}

func CountEntities[V models.Model](entityName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entityObject V
		var count int64

		forbiddenActionStr := fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead)

		userId, ok := before(w, r, entityName, nil)
		if !ok {
			http.Error(w, forbiddenActionStr, http.StatusForbidden)
			return
		}

		if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, entityName) {
			http.Error(w, forbiddenActionStr, http.StatusForbidden)
			return
		}

		query := database.GetDatabaseConnection().Model(&entityObject)
		countResult := query.Count(&count)

		if countResult.Error != nil {
			fmt.Printf("CountEntities error for %s entity", entityName)
			fmt.Print(countResult.Error)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		response := models.ModelCountResponse{
			Entity: entityName,
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
