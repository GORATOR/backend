package setup

import (
	"fmt"
	"strings"

	"github.com/GORATOR/backend/internal/models"
	"gorm.io/gorm"
)

// joinedChild describes one many2many join table linking rows of the
// deduplicated table (joinColumn) to their related rows (otherColumn). Join
// tables have a composite primary key, so re-pointing links with UPDATE could
// violate it: missing canonical links are inserted first, then links to the
// duplicates and the duplicates themselves are removed.
type joinedChild struct {
	joinTable   string
	joinColumn  string
	otherColumn string
}

// DeduplicateData merges rows left over from earlier, non-idempotent -s runs
// so that the unique indexes in SetupDatabase can be created over existing
// data and the seeded entities do not pile up: rows sharing the same logical
// key are collapsed into the oldest one, child references are re-pointed to
// it and the duplicates are removed. Projects with a clashing envelope_key
// keep all their data but receive a fresh key (their DSN was ambiguous
// anyway).
func DeduplicateData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Plain child references: safe to re-point with UPDATE.
		if err := dedupRows(tx, "event_common_sdks", []string{"name", "version"}, nil,
			"envelope_event_commons", "event_common_sdk_id"); err != nil {
			return err
		}
		if err := dedupRows(tx, "transactions", []string{"event_id"}, nil,
			"spans", "transaction_id"); err != nil {
			return err
		}
		if err := dedupRows(tx, "envelope_tags", []string{"name", "value"},
			[]joinedChild{
				{joinTable: "eetags_eecommon", joinColumn: "envelope_tag_id", otherColumn: "envelope_event_common_id"},
			}, "", ""); err != nil {
			return err
		}
		// Seeded organization/team copies: merge their join-table links, then
		// remove the duplicates. Real entities with distinct names are untouched.
		if err := dedupRows(tx, "organizations", []string{"name"},
			[]joinedChild{
				{joinTable: "org_teams", joinColumn: "organization_id", otherColumn: "team_id"},
				{joinTable: "org_users", joinColumn: "organization_id", otherColumn: "user_id"},
			}, "", ""); err != nil {
			return err
		}
		if err := dedupRows(tx, "teams", []string{"name"},
			[]joinedChild{
				{joinTable: "org_teams", joinColumn: "team_id", otherColumn: "organization_id"},
				{joinTable: "team_users", joinColumn: "team_id", otherColumn: "user_id"},
			},
			"projects", "team_id"); err != nil {
			return err
		}
		return dedupProjectEnvelopeKeys(tx)
	})
}

func prefixColumns(prefix string, keys []string) string {
	prefixed := make([]string, len(keys))
	for i, key := range keys {
		prefixed[i] = prefix + "." + key
	}
	return strings.Join(prefixed, ", ")
}

func canonicalSubquery(table string, keys []string) string {
	keyList := strings.Join(keys, ", ")
	return fmt.Sprintf("SELECT %s, min(id) AS id FROM %s GROUP BY %s", keyList, table, keyList)
}

// dedupRows collapses duplicate rows of table identified by keys into the
// oldest row: join-table links are merged first, then plain child references
// (if childTable is not empty) are re-pointed, then the duplicates are
// removed. childTable/childColumn must be empty for join-table-only children.
func dedupRows(db *gorm.DB, table string, keys []string, children []joinedChild, childTable string, childColumn string) error {
	var err error
	var rePointed int64
	for _, child := range children {
		var inserted, removedLinks int64
		inserted, err = insertMissingCanonicalLinks(db, table, keys, child)
		if err != nil {
			return err
		}
		removedLinks, err = removeLinksToDuplicates(db, table, keys, child)
		if err != nil {
			return err
		}
		if inserted > 0 || removedLinks > 0 {
			fmt.Printf("Merged duplicate %s references in %s: inserted %d canonical links, removed %d stale links\n",
				table, child.joinTable, inserted, removedLinks)
		}
	}
	if childTable != "" {
		rePointed, err = repointChildReferences(db, table, keys, childTable, childColumn)
		if err != nil {
			return err
		}
	}
	removed, err := removeDuplicateRows(db, table, keys)
	if err != nil {
		return err
	}
	if rePointed > 0 || removed > 0 {
		fmt.Printf("Deduplicated %s: re-pointed %d %s references, removed %d duplicate rows\n",
			table, rePointed, childTable, removed)
	}
	return nil
}

func insertMissingCanonicalLinks(db *gorm.DB, table string, keys []string, child joinedChild) (int64, error) {
	sql := fmt.Sprintf(`
INSERT INTO %[1]s (%[2]s, %[3]s)
SELECT DISTINCT canon.id, j.%[3]s
FROM %[1]s j
JOIN %[4]s dup ON dup.id = j.%[2]s
JOIN LATERAL (
    SELECT min(id) AS id FROM %[4]s m WHERE (%[5]s) = (%[6]s)
) canon ON canon.id < dup.id
WHERE NOT EXISTS (
    SELECT 1 FROM %[1]s existing
    WHERE existing.%[2]s = canon.id AND existing.%[3]s = j.%[3]s
)`,
		child.joinTable, child.joinColumn, child.otherColumn, table,
		prefixColumns("m", keys), prefixColumns("dup", keys))
	result := db.Exec(sql)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert canonical links into %s: %w", child.joinTable, result.Error)
	}
	return result.RowsAffected, nil
}

func removeLinksToDuplicates(db *gorm.DB, table string, keys []string, child joinedChild) (int64, error) {
	sql := fmt.Sprintf(`
DELETE FROM %[1]s j
USING %[2]s dup
WHERE j.%[3]s = dup.id
  AND EXISTS (
    SELECT 1 FROM %[2]s m
    WHERE (%[4]s) = (%[5]s) AND m.id < dup.id
  )`,
		child.joinTable, table, child.joinColumn,
		prefixColumns("m", keys), prefixColumns("dup", keys))
	result := db.Exec(sql)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to remove links to duplicate %s rows from %s: %w",
			table, child.joinTable, result.Error)
	}
	return result.RowsAffected, nil
}

func repointChildReferences(db *gorm.DB, table string, keys []string, childTable string, childColumn string) (int64, error) {
	sql := fmt.Sprintf(`
UPDATE %[1]s child
SET %[2]s = canon.id
FROM (%[3]s) canon
JOIN %[4]s dup ON (%[5]s) = (%[6]s) AND dup.id > canon.id
WHERE child.%[2]s = dup.id`,
		childTable, childColumn, canonicalSubquery(table, keys), table,
		prefixColumns("dup", keys), prefixColumns("canon", keys))
	result := db.Exec(sql)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to re-point %s.%s to canonical %s rows: %w",
			childTable, childColumn, table, result.Error)
	}
	return result.RowsAffected, nil
}

func removeDuplicateRows(db *gorm.DB, table string, keys []string) (int64, error) {
	sql := fmt.Sprintf(`
DELETE FROM %[1]s dup
USING %[1]s m
WHERE (%[2]s) = (%[3]s) AND dup.id > m.id`,
		table, prefixColumns("dup", keys), prefixColumns("m", keys))
	result := db.Exec(sql)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete duplicate rows from %s: %w", table, result.Error)
	}
	return result.RowsAffected, nil
}

func dedupProjectEnvelopeKeys(db *gorm.DB) error {
	var dupProjects []models.Project
	dupResult := db.Raw(`
SELECT id, name FROM projects p
WHERE EXISTS (
    SELECT 1 FROM projects p2
    WHERE p2.envelope_key = p.envelope_key AND p2.id < p.id
)`).Scan(&dupProjects)
	if dupResult.Error != nil {
		return fmt.Errorf("failed to find projects with duplicate envelope_key: %w", dupResult.Error)
	}

	for i := range dupProjects {
		project := &dupProjects[i]
		result := db.Unscoped().Model(&models.Project{}).
			Where("id = ?", project.ID).
			Update("envelope_key", models.NewEnvelopeKey(project.Name))
		if result.Error != nil {
			return fmt.Errorf("failed to regenerate envelope_key for project %d: %w", project.ID, result.Error)
		}
	}
	if len(dupProjects) > 0 {
		fmt.Printf("Regenerated envelope_key for %d project(s) with duplicate keys\n", len(dupProjects))
	}
	return nil
}
