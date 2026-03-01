package setup

import (
	"fmt"

	"gorm.io/gorm"
)

func MigrateExtraField(db *gorm.DB) error {
	err := db.Exec(`
		ALTER TABLE envelope_event_commons
		ADD COLUMN IF NOT EXISTS extra_data JSONB
	`).Error
	if err != nil {
		return fmt.Errorf("failed to add extra_data column: %w", err)
	}
	return nil
}
