package crud

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/GORATOR/backend/internal/api"
)

func before(w http.ResponseWriter, r *http.Request, entity string, entityId *uint) bool {
	_, userId := api.IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, api.MessageUnauthorized, http.StatusUnauthorized)
		return false
	}
	if entityId != nil {
		id, err := strconv.Atoi(r.URL.Path[len("/"+entity+"/"):])
		fmt.Println(entityId)
		*entityId = uint(id)
		if err != nil {
			http.Error(w, "Invalid resource ID", http.StatusBadRequest)
			return false
		}
	}
	return true
}
