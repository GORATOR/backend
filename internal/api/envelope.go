package api

import (
	"fmt"
	"net/http"
	"strconv"
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

	postItems := tryParsePostItems(body)
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

	tags, err := service.ParseTags(postItems)
	if err != nil {
		utils.HttpReturnBadRequest(w)
		return
	}
	fmt.Println(tags)

	if commonRecord.EventId == "" {
		utils.HttpReturnBadRequest(w)
		return
	}

	err = commonRecord.TryExtractKey(r)
	if err != nil {
		utils.HttpReturnBadRequest(w)
		return
	}

	projectId := tryParseProjectId(r)
	if projectId == 0 {
		utils.HttpReturnBadRequest(w)
		return
	}

	project := models.Project{}
	projectResult := database.GetDatabaseConnection().
		Where("id = ? and active = true", projectId).
		First(&project)
	if projectResult.Error != nil {
		fmt.Print(projectResult.Error)
		utils.HttpReturnBadRequest(w)
		return
	}

	if project.EnvelopeKey != commonRecord.EnvelopeKey {
		fmt.Print(err)
		utils.HttpReturnBadRequest(w)
		return
	}

	commonRecord.ProjectID = project.ID

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

func tryParsePostItems(body []byte) []string {
	postItems := strings.Split(string(body), "\n")
	result := []string{}
	for _, item := range postItems {
		if len(item) > 0 {
			result = append(result, item)
		}
	}
	return result
}

func tryParseProjectId(r *http.Request) uint {
	projectId := r.PathValue("id")
	projectIdUint, err := strconv.ParseUint(projectId, 10, 32)
	if err != nil {
		return 0
	}
	return uint(projectIdUint)
}
