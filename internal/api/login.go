package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
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

	//todo: calc password hash

	var user models.User
	db := database.GetDatabaseConnection()
	searchResult := db.Where(&user, "username = ?", credReq.Username).Where("password = ?", credReq.Password).First(&user)
	if searchResult.Error != nil {
		//todo: fmt
		http.Error(w, "Not found", http.StatusNotFound)
	}

	//todo: create session

}
