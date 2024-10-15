package crud

import (
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
)

func Delete[V models.Entity](entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		if !before(w, r, entity, &id) {
			return
		}

		/*if !service.HasUserAccessToByUserId(id, models.ActionDelete, entityInterface) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}*/
		_, err := database.DisableRecord[V](id)
		if err != nil {
			http.Error(w, "DB error", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
