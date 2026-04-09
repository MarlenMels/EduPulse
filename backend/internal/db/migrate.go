package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(db *sql.DB) error {
	entries, err := fs.Glob(migrationsFS, "migrations/*.sql")
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
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}