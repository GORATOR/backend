package crud

import (
	"encoding/json"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
)

func Read[V models.Entity](entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		if !before(w, r, entity, &id) {
			return
		}

		/*if !service.HasUserAccessToByUserId(id, models.ActionRead, entityInterface) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}*/

		data, err := database.GetRecord[V](id)
		if err != nil {
			http.Error(w, "DB error", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(data)
		w.WriteHeader(http.StatusOK)

	}
}
