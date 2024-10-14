package crud

import (
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
)

func Delete(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		if !before(w, r, entity, &id) {
			return
		}

		/*if !service.HasUserAccessToByUserId(id, models.ActionDelete, entityInterface) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}*/

		var entity = utils.EntityNameToInterface(entity)

		switch entity {
		case models.UserEntityName:
			var user = models.User{}
			deleteFromDb(id, &user, w)
		case models.OrganizationEntityName:
			var org = models.Organization{}
			deleteFromDb(id, &org, w)
		case models.TeamEntityName:
			var team = models.Team{}
			deleteFromDb(id, &team, w)
		}
	}
}

func deleteFromDb(id uint, entity interface{}, w http.ResponseWriter) {
	if database.DisableRecord(id, entity) != nil {
		http.Error(w, "DB error", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
