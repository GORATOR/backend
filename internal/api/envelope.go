package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
)

func Envelope(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		envelopeBadRequest(w)
		return
	}
	bodyAsString := string(body)
	postItems := strings.Split(bodyAsString, "\n")
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

	utils.HttpReturnJson(
		w,
		models.EnvelopeResponse{
			Id: commonData.EventId,
		},
	)
}

func envelopeBadRequest(w http.ResponseWriter) {
	http.Error(w, "Invalid request body", http.StatusBadRequest)
}
