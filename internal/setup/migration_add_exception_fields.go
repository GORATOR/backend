package setup

import (
	"fmt"

	"gorm.io/gorm"
)

type ExceptionValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ExceptionWrapper struct {
	Exception interface{} `json:"exception"`
}

func MigrateExceptionFields(db *gorm.DB) error {
	err := db.Exec(`
		ALTER TABLE envelope_event_commons
		ADD COLUMN IF NOT EXISTS exception_type TEXT,
		ADD COLUMN IF NOT EXISTS exception_value TEXT,
		ADD COLUMN IF NOT EXISTS exception_data JSONB
	`).Error
	if err != nil {
		return fmt.Errorf("failed to add columns: %w", err)
	}
	err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_exception_type_value
		ON envelope_event_commons (exception_type, exception_value)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	return nil
}
