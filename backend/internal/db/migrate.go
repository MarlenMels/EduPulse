package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql migrations_pg/*.sql
var migrationsFS embed.FS

func Migrate(db *sql.DB, driver Driver) error {
	dir := "migrations"
	if driver == DriverPostgres {
		dir = "migrations_pg"
	}

	entries, err := fs.Glob(migrationsFS, dir+"/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(entries)

	for _, name := range entries {
		b, err := migrationsFS.ReadFile(name)
		if err != nil {
			return err
		}
		sqlText := strings.TrimSpace(string(b))
		if sqlText == "" {
			continue
		}
		if _, err := db.Exec(sqlText); err != nil {
			// SQLite ALTER TABLE ADD COLUMN does not support IF NOT EXISTS,
			// so the same migration applied twice yields "duplicate column".
			// Postgres uses IF NOT EXISTS and never produces this error.
			if strings.Contains(err.Error(), "duplicate column") {
				continue
			}
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}
