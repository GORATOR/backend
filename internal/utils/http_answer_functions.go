package utils

import (
	"encoding/json"
	"net/http"
)

func HttpReturnBadRequest(w http.ResponseWriter) {
	http.Error(w, "Invalid request body", http.StatusBadRequest)
}

func HttpReturnJson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
