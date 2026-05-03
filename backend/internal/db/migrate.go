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
		// Split on semicolons so multi-statement files apply correctly under
		// pgx (which silently truncates multi-statement Exec in some cases).
		// modernc/sqlite handles either form.
		for _, stmt := range splitSQL(string(b)) {
			if _, err := db.Exec(stmt); err != nil {
				// SQLite ALTER TABLE ADD COLUMN does not support IF NOT EXISTS,
				// so the same migration applied twice yields "duplicate column".
				if strings.Contains(err.Error(), "duplicate column") {
					continue
				}
				return fmt.Errorf("apply %s: %w\n--statement--\n%s", name, err, stmt)
			}
		}
	}
	return nil
}

// splitSQL splits a SQL file into individual statements separated by `;` at the
// top level. It strips line comments (`-- ...`) and ignores empty trailing
// fragments. It does NOT understand string literals or dollar-quoted blocks —
// that is fine for our migration files which use neither.
func splitSQL(s string) []string {
	var out []string
	var cur strings.Builder
	for _, line := range strings.Split(s, "\n") {
		// strip "-- comment" tail (only when not inside a string literal —
		// our migrations have no inline strings with `--`)
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		cur.WriteString(line)
		cur.WriteByte('\n')
		// flush whenever a statement-terminating ';' appears
		text := cur.String()
		for {
			idx := strings.Index(text, ";")
			if idx < 0 {
				break
			}
			stmt := strings.TrimSpace(text[:idx])
			if stmt != "" {
				out = append(out, stmt)
			}
			text = text[idx+1:]
		}
		cur.Reset()
		cur.WriteString(text)
	}
	tail := strings.TrimSpace(cur.String())
	if tail != "" {
		out = append(out, tail)
	}
	return out
}
