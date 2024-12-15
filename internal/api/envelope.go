package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

func Envelope(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := utils.GetBodyBytes(r)
	if err != nil {
		utils.HttpReturnBadRequest(w)
		return
	}

	postItems := isValidRequest(body)
	if len(postItems) != models.EnvelopeRequiredPostItems {
		utils.HttpReturnBadRequest(w)
		return
	}

	var commonRecord models.EnvelopeEventCommon
	err = service.ParseSDK(&commonRecord, postItems)
	if err != nil {
		utils.HttpReturnBadRequest(w)
		return
	}

	if commonRecord.EventId == "" {
		utils.HttpReturnBadRequest(w)
		return
	}

	err = commonRecord.TryExtractKeyFromDsn()
	if err != nil {
		utils.HttpReturnBadRequest(w)
		return
	}

	err = database.EnvelopeSaveData(&commonRecord, postItems)
	if err != nil {
		fmt.Println(err)
		utils.HttpReturnBadRequest(w)
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
