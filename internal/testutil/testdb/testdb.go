package testdb

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

// NewTestDB creates a new in-memory SQLite database and runs migrations
func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Find migrations directory relative to this file
	// Current file: internal/testutil/testdb/testdb.go
	_, filename, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(filename)
	migrationsDir := filepath.Join(basepath, "..", "..", "infrastructure", "persistence", "sqlite", "migrations")

	// Set goose to be quiet during tests
	goose.SetLogger(goose.NopLogger())

	// Set dialect
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("failed to set dialect: %v", err)
	}

	// Run migrations
	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations from %q: %v", migrationsDir, err)
	}

	// Tidy up after test - note: since it is in-memory, closing is enough
	t.Cleanup(func() {
		db.Close()
	})

	return db
}
