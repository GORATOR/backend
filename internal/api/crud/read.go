package crud

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

func Read[V models.Entity](entity string) http.HandlerFunc {
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

func getQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func tryGetRecords[V models.Entity](query *gorm.DB, w http.ResponseWriter) {
	var entities []V
	result := query.Find(&entities)
	if result.Error != nil {
		fmt.Print("GetUsers query.Find error", result.Error)
		w.WriteHeader(http.StatusBadRequest)
	}
	utils.HttpReturnJson(w, entities)
}

func parseOffsetAndLimit[V models.Entity](w http.ResponseWriter, r *http.Request, entity string) (*gorm.DB, error) {
	var entityObject V
	query := database.GetDatabaseConnection().Model(&entityObject)

	//todo: переделать
	userId, ok := before(w, r, entity, nil)
	if !ok {
		return nil, errors.New("")
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, entity) {
		http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
		return nil, errors.New("")
	}

	offset := getQueryParam(r, "offset")
	limit := getQueryParam(r, "limit")

	if limit != "" {
		limitInt, _ := strconv.Atoi(limit)
		query.Limit(limitInt)
	} else {
		query.Limit(defaultLimit)
	}
	if offset != "" {
		offsetInt, _ := strconv.Atoi(offset)
		query.Offset(offsetInt)
	} else {
		query.Offset(defaultOffset)
	}

	return query, nil
}

func GetUsers[V models.Entity](w http.ResponseWriter, r *http.Request) {

	query, err := parseOffsetAndLimit[V](w, r, models.UserEntityName)
	if err != nil {
		return
	}
	parseUsersQuery(query, r)
	//todo: обработка моделей перед сериализацией
	tryGetRecords[V](query, w)
}

func GetTeams[V models.Entity](w http.ResponseWriter, r *http.Request) {
	query, err := parseOffsetAndLimit[V](w, r, models.TeamEntityName)
	if err != nil {
		return
	}
	parseTeamsQuery(query, r)
	tryGetRecords[V](query, w)
}

func parseUsersQuery(query *gorm.DB, r *http.Request) {
	username := getQueryParam(r, "username")
	if username != "" {
		query.Where("username like ?", username+"%")
	}
}

func parseTeamsQuery(query *gorm.DB, r *http.Request) {
	name := getQueryParam(r, "name")
	if name != "" {
		query.Where("name like ?", name+"%")
	}
}
