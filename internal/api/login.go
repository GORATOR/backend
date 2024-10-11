package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

var (
	sessionStore = service.NewRAMSessionStore()
)

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credReq models.CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&credReq); err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	salt := utils.StringFromEnv("GORATOR_SALT", "")
	if salt == "" {
		log.Printf("Empty env GORATOR_SALT")
	}

	hash := utils.HashPassword(credReq.Password, salt)

	var user models.User
	db := database.GetDatabaseConnection()
	searchResult := db.Where(&models.User{Username: credReq.Username, Password: hash, Active: true}).First(&user)
	if searchResult.Error != nil {
		log.Printf("No user for ")
		http.Error(w, "Not found", http.StatusNotFound)
	}

	session, err := sessionStore.CreateSession(int(user.ID))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: session.ID,
		Path:  "/",
	})
	resp := models.CredentialsResponse{
		User:      user,
		SessionId: session.ID,
	}
	json.NewEncoder(w).Encode(resp)
	w.WriteHeader(http.StatusOK)

}
