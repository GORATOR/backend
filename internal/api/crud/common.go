package crud

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/GORATOR/backend/internal/api"
	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

func before(w http.ResponseWriter, r *http.Request, entity string, entityId *uint) (int, bool) {
	_, userId := api.IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, api.MessageUnauthorized, http.StatusUnauthorized)
		return 0, false
	}
	if entityId != nil {
		id, err := strconv.Atoi(r.URL.Path[len("/"+entity+"/"):])
		fmt.Println(entityId)
		*entityId = uint(id)
		if err != nil {
			http.Error(w, "Invalid resource ID", http.StatusBadRequest)
			return userId, false
		}
	}
	return userId, true
}

func tryBuildReadQuery[V models.Entity](w http.ResponseWriter, r *http.Request, entity string) (*gorm.DB, error) {
	var entityObject V
	query := database.GetDatabaseConnection().Model(&entityObject)

	forbiddenActionStr := fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead)

	userId, ok := before(w, r, entity, nil)
	if !ok {
		http.Error(w, forbiddenActionStr, http.StatusForbidden)
		return nil, errors.New("can't get userId from session")
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, entity) {
		http.Error(w, forbiddenActionStr, http.StatusForbidden)
		return nil, fmt.Errorf("user with userId=%d hasn't access to %s entity", userId, entity)
	}

	query.Where("active = true")

	offset := utils.GetQueryParam(r, "offset")
	limit := utils.GetQueryParam(r, "limit")

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
