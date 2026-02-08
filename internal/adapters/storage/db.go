package storage

import (
	"fmt"
	"os"

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

// runMigrations executes the SQL schema file.
func runMigrations(db *sqlx.DB) error {
	// Read the schema file
	schema, err := os.ReadFile("migrations/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute the schema
	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}
