package crud

import (
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
)

func Delete(m models.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id uint
		userId, ok := before(w, r, m, &id)
		if !ok {
			return
		}

		if !service.HasUserAccessToByUserId(uint(userId), models.ActionDelete, m.GetName()) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
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
