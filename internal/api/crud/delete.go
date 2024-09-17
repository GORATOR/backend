package crud

import (
	"fmt"
	"net/http"
	"strconv"
)

func Delete(entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Path[len("/"+entity+"/"):])
		if err != nil {
			http.Error(w, "Invalid resource ID", http.StatusBadRequest)
			return
		}
		fmt.Println("crud read called", id)
	}
}
