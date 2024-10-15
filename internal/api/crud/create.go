package crud

import (
	"net/http"

	"github.com/GORATOR/backend/internal/models"
)

func Create[V models.Entity](entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !before(w, r, entity, nil) {
			return
		}
		//decode input
		//filter fields

	}
}
