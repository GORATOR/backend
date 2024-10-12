package service

import "github.com/GORATOR/backend/internal/models"

func HasUserAccessToByModel(user *models.User, action models.RuleAction, entity interface{}) bool {
	return HasUserAccessToByUserId(user.ID, action, entity)
}

func HasUserAccessToByUserId(userId uint, action models.RuleAction, entity interface{}) bool {
	return true
}
