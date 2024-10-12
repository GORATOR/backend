package crud

import (
	"encoding/json"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
)

func Read(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		if !before(w, r, entity, &id) {
			return
		}

		/*if !service.HasUserAccessToByUserId(id, models.ActionRead, entityInterface) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}*/

		switch entity {
		case models.UserEntityName:
			var user = models.User{}
			readFromDb(id, &user, w)
		case models.OrganizationEntityName:
			var org = models.Organization{}
			readFromDb(id, &org, w)
		case models.TeamEntityName:
			var team = models.Team{}
			readFromDb(id, &team, w)
		}

	}
}

func readFromDb(id uint, entity interface{}, w http.ResponseWriter) {
	db := database.GetDatabaseConnection()
	result := db.Where("id = ?", id).First(entity)
	if result.Error != nil {
		http.Error(w, "DB error", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(entity)
	w.WriteHeader(http.StatusOK)
}
