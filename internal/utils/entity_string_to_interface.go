package utils

import "github.com/GORATOR/backend/internal/models"

func EntityNameToInterface(entityName string) interface{} {
	switch entityName {
	case models.UserEntityName:
		return models.User{}
	case models.TeamEntityName:
		return models.Team{}
	case models.OrganizationEntityName:
		return models.Organization{}
	default:
		return nil
	}
}
