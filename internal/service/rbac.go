package service

import (
	"fmt"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
)

var (
	sqlHasUserAccess = `select distinct ru.id, ru.allowed from rules ru
	left join roles ro on ru.role_id = ro.id
	left join role_users rou on rou.role_id = ro.id and rou.user_id = ?
	where ru."table" = ? and ru."action" = ? and rou.user_id is not null`

	sqlHasUserRole = `select count(*) from role_users rou
	left join roles ro on rou.role_id = ro.id
	where rou.user_id = ? and ro.name = ?`
)

func HasUserAccessToByModel(user *models.User, action models.RuleAction, entity interface{}) bool {
	return HasUserAccessToByUserId(user.ID, action, entity)
}

func HasUserAccessToByUserId(userId uint, action models.RuleAction, entity interface{}) bool {
	tableName := fmt.Sprintf("%ss", entity)
	var rule models.Rule
	db := database.GetDatabaseConnection()
	result := db.Raw(sqlHasUserAccess, userId, tableName, string(action)).Find(&rule)
	if result.Error != nil || result.RowsAffected != 1 {
		//todo: log error
		return false
	}
	return rule.Allowed
}

func HasUserRole(userId uint, roleName string) bool {
	var count int64
	db := database.GetDatabaseConnection()
	result := db.Raw(sqlHasUserRole, userId, roleName).Scan(&count)
	if result.Error != nil {
		//todo: log error
		return false
	}
	return count > 0
}
