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

	fmt.Printf("[envelope] items=%d\n", len(postItems))
	for i, item := range postItems {
		preview := item
		if len(preview) > 120 {
			preview = preview[:120] + "..."
		}
		fmt.Printf("[envelope] [%d] %s\n", i, preview)
	}

	// Check transaction first — envelopes may contain both transaction + client_report
	if models.IsTransaction(postItems) {
		handleTransaction(w, r, postItems)
		return
	}

	// Check if this is a client report
	if models.IsClientReport(postItems) {
		handleClientReport(w, r, postItems)
		return
	}

	fmt.Printf("[envelope] no handler matched (isClientReport=false, isTransaction=false), items=%d\n", len(postItems))

	// Handle regular envelope (existing logic)
	if len(postItems) != models.EnvelopeRequiredPostItems {
		fmt.Printf("[envelope] bad item count: got %d, want %d\n", len(postItems), models.EnvelopeRequiredPostItems)
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

	err = service.ParseException(&commonRecord, postItems)
	if err != nil {
		utils.HttpReturnBadRequest(w)
		return
	}

	err = service.ParseExtra(&commonRecord, postItems)
	if err != nil {
		utils.HttpReturnBadRequest(w)
		return
	}

	err = service.ParseMessage(&commonRecord, postItems)
	if err != nil {
		utils.HttpReturnBadRequest(w)
		return
	}

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

	commonRecord.ProjectID = &project.ID

	err = database.EnvelopeSaveData(&commonRecord, postItems, &tags)
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

func handleClientReport(w http.ResponseWriter, r *http.Request, postItems []string) {
	fmt.Println("Processing client report...")

	clientReport, err := service.ParseClientReport(postItems)
	if err != nil {
		fmt.Printf("Failed to parse client report: %v\n", err)
		utils.HttpReturnBadRequest(w)
		return
	}

	// Extract project key from request (same as regular envelope)
	var dummyRecord models.EnvelopeEventCommon
	err = dummyRecord.TryExtractKey(r)
	if err != nil {
		fmt.Printf("Failed to extract key from client report: %v\n", err)
		utils.HttpReturnBadRequest(w)
		return
	}

	clientReport.EnvelopeKey = dummyRecord.EnvelopeKey

	projectId := tryParseProjectId(r)
	if projectId == 0 {
		utils.HttpReturnBadRequest(w)
		return
	}

	// Verify project exists and key matches
	project := models.Project{}
	projectResult := database.GetDatabaseConnection().
		Where("id = ? and active = true", projectId).
		First(&project)
	if projectResult.Error != nil {
		fmt.Printf("Project not found: %v\n", projectResult.Error)
		utils.HttpReturnBadRequest(w)
		return
	}

	if project.EnvelopeKey != clientReport.EnvelopeKey {
		fmt.Printf("Project key mismatch: expected %s, got %s\n", project.EnvelopeKey, clientReport.EnvelopeKey)
		utils.HttpReturnBadRequest(w)
		return
	}

	clientReport.ProjectID = project.ID

	err = database.ClientReportSaveData(clientReport)
	if err != nil {
		fmt.Printf("Failed to save client report: %v\n", err)
		utils.HttpReturnBadRequest(w)
		return
	}

	// Return empty response (like Sentry does for client reports)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func handleTransaction(w http.ResponseWriter, r *http.Request, postItems []string) {
	fmt.Println("Processing transaction...")

	tx, err := service.ParseTransaction(postItems)
	if err != nil {
		fmt.Printf("Failed to parse transaction: %v\n", err)
		utils.HttpReturnBadRequest(w)
		return
	}

	// Reuse existing key extraction via EnvelopeEventCommon helper
	var dummy models.EnvelopeEventCommon
	if err := dummy.TryExtractKey(r); err != nil {
		fmt.Printf("Failed to extract key from transaction: %v\n", err)
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
		fmt.Printf("Project not found: %v\n", projectResult.Error)
		utils.HttpReturnBadRequest(w)
		return
	}

	if project.EnvelopeKey != dummy.EnvelopeKey {
		fmt.Printf("Project key mismatch for transaction\n")
		utils.HttpReturnBadRequest(w)
		return
	}

	tx.ProjectID = &project.ID

	if err := database.TransactionSaveData(tx); err != nil {
		fmt.Printf("Failed to save transaction: %v\n", err)
		utils.HttpReturnBadRequest(w)
		return
	}

	// Some SDKs bundle a client_report in the same envelope — save it too
	if models.IsClientReport(postItems) {
		if cr, err := service.ParseClientReport(postItems); err == nil {
			cr.EnvelopeKey = dummy.EnvelopeKey
			cr.ProjectID = project.ID
			database.ClientReportSaveData(cr)
		}
	}

	utils.HttpReturnJson(w, models.EnvelopeResponse{Id: tx.EventId})
}

func tryParseProjectId(r *http.Request) uint {
	projectId := r.PathValue("id")
	projectIdUint, err := strconv.ParseUint(projectId, 10, 32)
	if err != nil {
		return 0
	}
	return uint(projectIdUint)
}
