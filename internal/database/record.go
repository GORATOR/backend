package database

import "github.com/GORATOR/backend/internal/models"

const (
	activeRecordWhere = "id = ? and active = true"
)

func GetRecord(id uint, m models.Model) (*models.Model, error) {
	result := postgresConnection.Where(activeRecordWhere, id).First(&m)
	return &m, result.Error
}

func DisableRecord(id uint, m models.Model) (*models.Model, error) {
	result := postgresConnection.Model(&m).Where(activeRecordWhere, id).Update("active", false)
	return &m, result.Error
}

func EnableRecord(id uint, m models.Model) (*models.Model, error) {
	result := postgresConnection.Model(&m).Where("id = ? and active = false", id).Update("active", true)
	return &m, result.Error
}
