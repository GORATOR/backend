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

func tryGetRecords[V models.Entity](query *gorm.DB, entities *[]V) error {
	result := query.Find(&entities)
	if result.Error != nil {
		fmt.Print("tryGetRecords query.Find error", result.Error)
		return result.Error
	}
	return nil
}

func ReadUsers(w http.ResponseWriter, r *http.Request) {

	query, err := buildReadQuery[models.User](w, r, models.UserEntityName)
	if err != nil {
		return
	}
	var entities []models.User
	parseUsersQuery(query, r)
	result := query.Select("ID", "CreatedAt", "UpdatedAt", "Username", "Email", "Avatar", "Active").Find(&entities)

	if result.Error != nil {
		fmt.Print("ReadUsers query.Find error", result.Error)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	utils.HttpReturnJson(w, entities)
}

func ReadTeams(w http.ResponseWriter, r *http.Request) {
	query, err := buildReadQuery[models.Team](w, r, models.TeamEntityName)
	if err != nil {
		return
	}
	var entities []models.Team
	parseTeamsQuery(query, r)
	err = tryGetRecords(query, &entities)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	utils.HttpReturnJson(w, entities)
}

func parseUsersQuery(query *gorm.DB, r *http.Request) {
	username := utils.GetQueryParam(r, "username")
	if username != "" {
		query.Where("username like ?", username+"%")
	}
}

func parseTeamsQuery(query *gorm.DB, r *http.Request) {
	name := utils.GetQueryParam(r, "name")
	if name != "" {
		query.Where("name like ?", name+"%")
	}
}
