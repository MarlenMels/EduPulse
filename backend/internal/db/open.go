package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

type Driver string

const (
	DriverSQLite   Driver = "sqlite"
	DriverPostgres Driver = "postgres"
)

type Config struct {
	URL    string // when non-empty and looks like postgres://, use Postgres
	Path   string // SQLite file path (used when URL is empty)
	Driver Driver // populated by Open
}

func Open(cfg *Config) (*sql.DB, error) {
	if isPostgresURL(cfg.URL) {
		cfg.Driver = DriverPostgres
		return openPostgres(cfg.URL)
	}
	cfg.Driver = DriverSQLite
	return openSQLite(cfg.Path)
}

func isPostgresURL(u string) bool {
	u = strings.ToLower(strings.TrimSpace(u))
	return strings.HasPrefix(u, "postgres://") || strings.HasPrefix(u, "postgresql://")
}

func openSQLite(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_busy_timeout=5000", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer; keep it simple

	ctx, cancel := contextWithTimeout(5 * time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func openPostgres(url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := contextWithTimeout(10 * time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// OpenSQLite is kept as a thin wrapper for compatibility with existing callers/tests.
func OpenSQLite(path string) (*sql.DB, error) { return openSQLite(path) }
