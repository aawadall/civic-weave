package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// RunMigrations executes all pending migrations
func Migrate(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Load migration files
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if _, exists := appliedMigrations[migration.Version]; !exists {
			log.Printf("Applying migration %d: %s", migration.Version, migration.Name)
			
			if err := applyMigration(db, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func loadMigrations() ([]Migration, error) {
	migrationsDir := "migrations"
	
	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Println("No migrations directory found, skipping migrations")
		return []Migration{}, nil
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		migration, err := parseMigrationFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration %s: %w", file.Name(), err)
		}

		migrations = append(migrations, migration)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigrationFile(filepath string) (Migration, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return Migration{}, err
	}

	// Extract version from filename (e.g., "001_initial.sql" -> version 1)
	filename := filepath[strings.LastIndex(filepath, "/")+1:]
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	var version int
	if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
		return Migration{}, fmt.Errorf("invalid version in filename: %s", filename)
	}

	// Extract name (everything after version and underscore)
	name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

	// Split content by -- DOWN separator
	contentStr := string(content)
	parts = strings.Split(contentStr, "-- DOWN")
	
	var up, down string
	if len(parts) >= 1 {
		up = strings.TrimSpace(parts[0])
		// Remove any -- UP comment
		up = strings.TrimPrefix(up, "-- UP")
		up = strings.TrimSpace(up)
	}
	if len(parts) >= 2 {
		down = strings.TrimSpace(parts[1])
	}

	return Migration{
		Version: version,
		Name:    name,
		Up:      up,
		Down:    down,
	}, nil
}

func applyMigration(db *sql.DB, migration Migration) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(migration.Up); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration as applied
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}
