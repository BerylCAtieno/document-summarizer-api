package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// ensureDir ensures the parent directory of the DB file exists
func ensureDir(dbFile string) error {
	dir := filepath.Dir(dbFile)
	return os.MkdirAll(dir, 0755)
}

// NewSQLiteDB creates a new SQLite connection
func NewSQLiteDB(dbFile string) (*sqlx.DB, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute database path: %w", err)
	}

	// Ensure folder exists
	if err := ensureDir(absPath); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Connect
	db, err := sqlx.Connect("sqlite", absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return db, nil
}

// RunMigrations runs SQLite migrations
func RunMigrations(dbFile string) error {
	// Absolute DB path
	absDB, err := filepath.Abs(dbFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute DB path: %w", err)
	}

	// Ensure directory exists
	if err := ensureDir(absDB); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sqlx.Connect("sqlite", absDB)
	if err != nil {
		return fmt.Errorf("failed to connect for migrations: %w", err)
	}
	defer db.Close()

	driver, err := sqlite.WithInstance(db.DB, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Absolute migration directory
	migrationsPath, err := filepath.Abs("internal/db/migrations")
	if err != nil {
		return fmt.Errorf("failed to get migration path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"sqlite",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
