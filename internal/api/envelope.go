package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
)

func Envelope(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := utils.GetBodyBytes(r)
	if err != nil {
		envelopeBadRequest(w)
		return
	}

	postItems := isValidRequest(body)
	if len(postItems) != 3 {
		envelopeBadRequest(w)
		return
	}

	var commonRecord models.EnvelopeEventCommon
	if err := json.Unmarshal([]byte(postItems[0]), &commonRecord); err != nil {
		envelopeBadRequest(w)
		return
	}

	if commonRecord.EventId == "" {
		envelopeBadRequest(w)
		return
	}

	err = database.EnvelopeSaveData(&commonRecord, postItems)
	if err != nil {
		fmt.Println(err)
		envelopeBadRequest(w)
		return
	}

	utils.HttpReturnJson(
		w,
		models.EnvelopeResponse{
			Id: commonRecord.EventId,
		},
	)
}

func isValidRequest(body []byte) []string {
	postItems := strings.Split(string(body), "\n")
	result := []string{}
	for _, item := range postItems {
		if len(item) > 0 {
			result = append(result, item)
		}
	}
	return result
}

func envelopeBadRequest(w http.ResponseWriter) {
	http.Error(w, "Invalid request body", http.StatusBadRequest)
}
