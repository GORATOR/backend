package crud

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/GORATOR/backend/internal/api"
)

func Read(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, userId := api.IsAuthorized(r)
		if !(userId > 0) {
			http.Error(w, api.MessageUnauthorized, http.StatusUnauthorized)
			return
		}
		id, err := strconv.Atoi(r.URL.Path[len("/"+entity+"/"):])
		if err != nil {
			http.Error(w, "Invalid resource ID", http.StatusBadRequest)
			return
		}
		fmt.Println("crud read called", id)
	}
}
