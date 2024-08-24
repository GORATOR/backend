package api

import (
	"encoding/json"
	"fmt"
	"io"
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		envelopeBadRequest(w)
		return
	}

	postItems := isValidRequest(body)
	if len(postItems) != 3 {
		envelopeBadRequest(w)
		return
	}

	//todo: process every item of postItems

	var commonData models.EnvelopeRequestEventCommon
	if err := json.Unmarshal([]byte(postItems[0]), &commonData); err != nil {
		envelopeBadRequest(w)
		return
	}

	if commonData.EventId == "" {
		envelopeBadRequest(w)
		return
	}

	err = database.EnvelopeSaveData(&commonData, postItems)
	if err != nil {
		fmt.Println(err)
		envelopeBadRequest(w)
		return
	}

	utils.HttpReturnJson(
		w,
		models.EnvelopeResponse{
			Id: commonData.EventId,
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
