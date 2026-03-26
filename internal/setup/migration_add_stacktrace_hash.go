package setup

import (
	"fmt"

	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/valyala/fastjson"
	"gorm.io/gorm"
)

func MigrateStacktraceHash(db *gorm.DB) error {
	err := db.Exec(`
		ALTER TABLE envelope_event_commons
		ADD COLUMN IF NOT EXISTS stacktrace_hash TEXT
	`).Error
	if err != nil {
		return fmt.Errorf("failed to add stacktrace_hash column: %w", err)
	}
	err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_stacktrace_hash
		ON envelope_event_commons (stacktrace_hash)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create stacktrace_hash index: %w", err)
	}

	var records []models.EnvelopeEventCommon
	if err := db.Where("exception_data IS NOT NULL AND (stacktrace_hash IS NULL OR stacktrace_hash = '')").
		Select("id, exception_data").
		Find(&records).Error; err != nil {
		return fmt.Errorf("failed to load records for backfill: %w", err)
	}

	var p fastjson.Parser
	for _, rec := range records {
		if rec.ExceptionData == nil {
			continue
		}
		v, err := p.Parse(*rec.ExceptionData)
		if err != nil {
			continue
		}
		hash := service.ComputeStacktraceHash(v)
		if hash == "" {
			continue
		}
		db.Model(&models.EnvelopeEventCommon{}).Where("id = ?", rec.ID).Update("stacktrace_hash", hash)
	}

	fmt.Printf("Backfilled stacktrace_hash for %d records\n", len(records))
	return nil
}
