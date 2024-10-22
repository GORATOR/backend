package api

import (
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
)

func UserCurrent(w http.ResponseWriter, r *http.Request) {
	_, userId := IsAuthorized(r)
	if !(userId > 0) {
		http.Error(w, MessageUnauthorized, http.StatusUnauthorized)
		return
	}

	user, err := database.GetRecord[models.User](uint(userId))
	if err != nil {
		fmt.Print("database.GetRecord[models.User] error", err)
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}
	utils.HttpReturnJson(w, user)
}
