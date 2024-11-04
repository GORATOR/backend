package crud

import (
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
)

func Delete[V models.Model](entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		userId, ok := before(w, r, entity, &id)
		if !ok {
			return
		}

		if !service.HasUserAccessToByUserId(uint(userId), models.ActionDelete, entity) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}
		_, err := database.DisableRecord[V](id)
		if err != nil {
			http.Error(w, "DB error", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
