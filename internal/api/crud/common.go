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

func before(w http.ResponseWriter, r *http.Request, m models.ReadableModel, entityId *uint) (int, bool) {
	_, userId := api.IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, api.MessageUnauthorized, http.StatusUnauthorized)
		return 0, false
	}
	if entityId != nil {
		fmt.Println(entityId)
		id, err := extractIdFromUrl(m, r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return userId, false
		}
		*entityId = id
	}
	return userId, true
}

func extractIdFromUrl(m models.ReadableModel, path string) (uint, error) {
	entity := m.GetName()
	searchStr := "/" + entity + "/"
	if len(path) < len(searchStr) {
		for _, alias := range m.GetAliases() {
			searchStr = "/" + alias + "/"
			if len(path) < len(searchStr) {
				continue
			}
			id, err := strconv.Atoi(path[len("/"+alias+"/"):])
			if err == nil {
				return uint(id), nil
			}
		}
		return 0, fmt.Errorf("")
	}
	id, err := strconv.Atoi(path[len(searchStr):])
	if err != nil {
		for _, alias := range m.GetAliases() {
			searchStr = "/" + alias + "/"
			if len(path) < len(searchStr) {
				continue
			}
			id, err = strconv.Atoi(path[len("/"+alias+"/"):])
			if err == nil {
				return uint(id), nil
			}
		}
	}
	return uint(id), fmt.Errorf("")
}

func tryBuildReadQuery(w http.ResponseWriter, r *http.Request, m models.ReadableModel, ignoreActive bool) (*gorm.DB, error) {
	query := database.GetDatabaseConnection().Model(&m)

	forbiddenActionStr := fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead)

	userId, ok := before(w, r, m, nil)
	if !ok {
		http.Error(w, forbiddenActionStr, http.StatusForbidden)
		return nil, errors.New("can't get userId from session")
	}

	if !service.HasUserAccessToByUserId(uint(userId), models.ActionRead, m.GetName()) {
		http.Error(w, forbiddenActionStr, http.StatusForbidden)
		return nil, fmt.Errorf("user with userId=%d hasn't access to %s entity", userId, m.GetName())
	}

	if ignoreActive != true {
		query.Where("active = true")
	}

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
