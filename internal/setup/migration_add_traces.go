package setup

import (
	"fmt"

	"gorm.io/gorm"
)

func MigrateTraces(db *gorm.DB) error {
	indexes := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS unique_transaction_event_id ON transactions (event_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transaction_trace_id ON transactions (trace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transaction_parent_span_id ON transactions (parent_span_id)`,
		`CREATE INDEX IF NOT EXISTS idx_span_trace_id ON spans (trace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_span_parent_span_id ON spans (parent_span_id)`,
		`CREATE INDEX IF NOT EXISTS idx_span_span_id ON spans (span_id)`,
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index (%s): %w", idx, err)
		}
	}

	fmt.Println("Trace tables migration done")
	return nil
}
