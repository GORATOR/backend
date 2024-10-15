package crud

import (
	"net/http"

	"github.com/GORATOR/backend/internal/models"
)

func Update[V models.Entity](entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		if !before(w, r, entity, &id) {
			return
		}

		/*if !service.HasUserAccessToByUserId(id, models.ActionRead, entityInterface) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}*/

	}
}
