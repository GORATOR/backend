package service

import (
	"github.com/GORATOR/backend/internal/models"
)

func HasUserAccessTo(user *models.User, entity string) bool {
	return false
}
