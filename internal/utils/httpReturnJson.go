package utils

import (
	"encoding/json"
	"net/http"
)

func HttpReturnJson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
