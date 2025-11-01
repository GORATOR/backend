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

// UserRead handles reading a single user by ID
// Users can read themselves, admins can read anyone
func UserRead(m models.ReadableModel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		userId, ok := before(w, r, m, &id)
		if !ok {
			return
		}

		// Allow if reading self OR if user is admin
		if id != uint(userId) && !service.HasUserRole(uint(userId), "admin") {
			http.Error(w, "Only users with admin role can read other users", http.StatusForbidden)
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

// UserReadEntities handles reading list of users
// Only admins can read user list
func UserReadEntities(m models.ReadableModel, ignoreActive bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := tryBuildUserReadQuery(w, r, m, ignoreActive)
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

// UserCountEntities handles counting users
// Only admins can count users
func UserCountEntities(m models.ReadableModel, ignoreActive bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := tryBuildUserReadQuery(w, r, m, ignoreActive)
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

// UserUpdate handles updating a user
// Users can update themselves, admins can update anyone
func UserUpdate(m models.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, ok := before(w, r, m, nil)
		if !ok {
			return
		}

		body, err := utils.GetBodyBytes(r)
		if err != nil {
			utils.HttpReturnBadRequest(w)
			return
		}

		err = json.Unmarshal(body, &m)
		if err != nil {
			fmt.Print("update json.Unmarshal error", err)
			utils.HttpReturnBadRequest(w)
			return
		}

		var input struct {
			ID uint `json:"ID"`
		}
		if err := json.Unmarshal(body, &input); err == nil {
			// Allow if editing self OR if user is admin
			if input.ID != uint(userId) && !service.HasUserRole(uint(userId), "admin") {
				http.Error(w, "Only users with admin role can edit other users", http.StatusForbidden)
				return
			}
		}

		db := database.GetDatabaseConnection()
		result, err := m.UpdateModel(body, uint(userId), db)
		if err != nil {
			utils.HttpReturnBadRequest(w)
			return
		}
		utils.HttpReturnJson(w, result)
	}
}

// UserCreate handles creating a new user
// Only admins can create users
func UserCreate(m models.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, ok := before(w, r, m, nil)
		if !ok {
			return
		}

		// Only admins can create users
		if !service.HasUserRole(uint(userId), "admin") {
			http.Error(w, "Only users with admin role can create users", http.StatusForbidden)
			return
		}

		body, err := utils.GetBodyBytes(r)
		if err != nil {
			utils.HttpReturnBadRequest(w)
			return
		}

		db := database.GetDatabaseConnection()
		result, err := m.CreateModel(body, uint(userId), db)
		if err != nil {
			utils.HttpReturnBadRequest(w)
			return
		}
		utils.HttpReturnJson(w, result)
	}
}

// UserDelete handles deleting (disabling) a user
// Only admins can delete users
func UserDelete(m models.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		userId, ok := before(w, r, m, &id)
		if !ok {
			return
		}

		// Only admins can delete users
		if !service.HasUserRole(uint(userId), "admin") {
			http.Error(w, "Only users with admin role can delete users", http.StatusForbidden)
			return
		}

		_, err := database.DisableRecord(id, m)
		if err != nil {
			http.Error(w, "DB error", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// tryBuildUserReadQuery builds query for user read operations
// Only admins can read user list
func tryBuildUserReadQuery(w http.ResponseWriter, r *http.Request, m models.ReadableModel, ignoreActive bool) (*gorm.DB, error) {
	query := database.GetDatabaseConnection().Model(&m)

	userId, ok := before(w, r, m, nil)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return nil, errors.New("can't get userId from session")
	}

	// Only admins can read user list
	if !service.HasUserRole(uint(userId), "admin") {
		http.Error(w, "Only users with admin role can read user list", http.StatusForbidden)
		return nil, fmt.Errorf("user with userId=%d is not admin", userId)
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
