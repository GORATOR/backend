package database

import (
	"errors"

	"github.com/GORATOR/backend/internal/models"
	"gorm.io/gorm"
)

func TransactionSaveData(tx *models.Transaction) error {
	// Idempotent: skip duplicates on SDK retries (unique event_id)
	var existing models.Transaction
	result := postgresConnection.Where("event_id = ?", tx.EventId).First(&existing)
	if result.Error == nil {
		return nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// GORM cascades Spans automatically via foreignKey:TransactionID
	return postgresConnection.Create(tx).Error
}
