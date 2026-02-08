package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database connection and runs migrations.
func InitDB(dbPath string) (*sqlx.DB, error) {
	// Create the database connection
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite works best with a single connection

	// Run migrations
	err = runMigrations(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations discovers and executes all SQL migration files in order.
func runMigrations(db *sqlx.DB) error {
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return fmt.Errorf("failed to discover migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		schema, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		_, err = db.Exec(string(schema))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}
