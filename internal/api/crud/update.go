package crud

import (
	"net/http"

	"github.com/GORATOR/backend/internal/api"
)

func Update(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, userId := api.IsAuthorized(r)
		if !(userId > 0) {
			http.Error(w, api.MessageUnauthorized, http.StatusUnauthorized)
			return
		}
	}
}
