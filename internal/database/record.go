package database

import "github.com/GORATOR/backend/internal/models"

const (
	activeRecordWhere = "id = ? and active = true"
)

func GetRecord[V models.Entity](id uint) (*V, error) {
	var e V
	result := postgresConnection.Where(activeRecordWhere, id).First(&e)
	return &e, result.Error
}

func DisableRecord[V models.Entity](id uint) (*V, error) {
	var e V
	result := postgresConnection.Model(&e).Where(activeRecordWhere, id).Update("active", false)
	return &e, result.Error
}

func EnableRecord[V models.Entity](id uint) (*V, error) {
	var e V
	result := postgresConnection.Model(&e).Where("id = ? and active = false", id).Update("active", true)
	return &e, result.Error
}
