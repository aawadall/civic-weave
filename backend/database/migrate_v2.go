package database

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	Version         string
	Name            string
	Status          string
	ExecutionTimeMs int
	Error           error
}

// MigrationOptions represents options for migration operations
type MigrationOptions struct {
	DryRun           bool
	FailOnIncompatible bool
	MaxVersion       string
	AutoApprove     bool
	RuntimeVersion   string
}

// MigrateV2 runs the enhanced migration system
func MigrateV2(db *sql.DB, options *MigrationOptions) error {
	if options == nil {
		options = &MigrationOptions{}
	}

	// Create migrations table if it doesn't exist
	if err := createMigrationsTableV2(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Load migration registry
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return fmt.Errorf("failed to load migration registry: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrationsV2(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedVersions := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	// Get all available migrations sorted by version
	allMigrations := registry.GetSortedMigrations()

	// Find pending migrations
	var pendingMigrations []*MigrationMetadata
	for _, migration := range allMigrations {
		if !appliedVersions[migration.Version] {
			// Check if we should apply this migration based on MaxVersion
			if options.MaxVersion != "" {
				if migration.Version > options.MaxVersion {
					continue
				}
			}
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) == 0 {
		log.Println("No pending migrations found")
		return nil
	}

	// Apply pending migrations
	for _, migration := range pendingMigrations {
		// Check dependencies
		if err := registry.ValidateDependencies(migration.Version, appliedVersions); err != nil {
			return fmt.Errorf("dependency check failed for migration %s: %w", migration.Version, err)
		}

		// Check compatibility if runtime version is provided
		if options.RuntimeVersion != "" {
			compat, err := registry.CheckCompatibility(migration.Version, options.RuntimeVersion)
			if err != nil {
				return fmt.Errorf("compatibility check failed for migration %s: %w", migration.Version, err)
			}

			if !compat.IsCompatible {
				if options.FailOnIncompatible {
					return fmt.Errorf("migration %s is incompatible: %s", migration.Version, compat.Message)
				}
				log.Printf("Warning: migration %s compatibility issue: %s", migration.Version, compat.Message)
			}
		}

		// Apply migration
		result, err := applyMigrationV2(db, migration, options)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		if result.Status == "applied" {
			appliedVersions[migration.Version] = true
			log.Printf("✅ Applied migration %s (%s) in %dms", 
				migration.Version, migration.Name, result.ExecutionTimeMs)
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}

// RollbackV2 rolls back migrations to a specific version
func RollbackV2(db *sql.DB, targetVersion string, options *MigrationOptions) error {
	if options == nil {
		options = &MigrationOptions{}
	}

	// Create migrations table if it doesn't exist
	if err := createMigrationsTableV2(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Load migration registry
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return fmt.Errorf("failed to load migration registry: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrationsV2(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Find migrations to rollback (in reverse order)
	var rollbackMigrations []*MigrationMetadata
	for i := len(appliedMigrations) - 1; i >= 0; i-- {
		migration := appliedMigrations[i]
		if migration.Version <= targetVersion {
			break
		}

		metadata, exists := registry.GetMigration(migration.Version)
		if !exists {
			log.Printf("Warning: metadata not found for migration %s, skipping rollback", migration.Version)
			continue
		}

		rollbackMigrations = append(rollbackMigrations, metadata)
	}

	if len(rollbackMigrations) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Rollback migrations
	for _, migration := range rollbackMigrations {
		result, err := rollbackMigrationV2(db, migration, options)
		if err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
		}

		if result.Status == "rolled_back" {
			log.Printf("✅ Rolled back migration %s (%s) in %dms", 
				migration.Version, migration.Name, result.ExecutionTimeMs)
		}
	}

	log.Println("Rollback completed successfully")
	return nil
}

// createMigrationsTableV2 creates the enhanced migrations table
func createMigrationsTableV2(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations_v2 (
			version VARCHAR(20) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			checksum VARCHAR(64) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			runtime_version VARCHAR(20),
			execution_time_ms INTEGER,
			status VARCHAR(20) DEFAULT 'applied'
		);
	`
	
	_, err := db.Exec(query)
	return err
}

// applyMigrationV2 applies a single migration
func applyMigrationV2(db *sql.DB, migration *MigrationMetadata, options *MigrationOptions) (*MigrationResult, error) {
	startTime := time.Now()

	// Get migration file paths
	upPath, _ := GetMigrationPath("migrations_v2", migration.Version)
	if upPath == "" {
		return nil, fmt.Errorf("migration file not found for version %s", migration.Version)
	}

	// Read migration file
	upSQL, err := os.ReadFile(upPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file: %w", err)
	}

	// Calculate checksum
	checksum := calculateChecksum(upSQL)

	// Check if migration was already applied with same checksum
	var existingChecksum string
	err = db.QueryRow("SELECT checksum FROM schema_migrations_v2 WHERE version = $1", migration.Version).Scan(&existingChecksum)
	if err == nil && existingChecksum == checksum {
		log.Printf("Migration %s already applied with same checksum, skipping", migration.Version)
		return &MigrationResult{
			Version: migration.Version,
			Name:    migration.Name,
			Status:  "skipped",
		}, nil
	}

	// If checksum differs, this is an error (migration was modified)
	if err == nil && existingChecksum != checksum {
		return nil, fmt.Errorf("migration %s was modified after being applied (checksum mismatch)", migration.Version)
	}

	if options.DryRun {
		log.Printf("DRY RUN: Would apply migration %s (%s)", migration.Version, migration.Name)
		return &MigrationResult{
			Version: migration.Version,
			Name:    migration.Name,
			Status:  "dry_run",
		}, nil
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(string(upSQL)); err != nil {
		return nil, fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration as applied
	executionTime := int(time.Since(startTime).Milliseconds())
	_, err = tx.Exec(`
		INSERT INTO schema_migrations_v2 (version, name, checksum, runtime_version, execution_time_ms, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (version) DO UPDATE SET
			checksum = EXCLUDED.checksum,
			runtime_version = EXCLUDED.runtime_version,
			execution_time_ms = EXCLUDED.execution_time_ms,
			status = EXCLUDED.status,
			applied_at = CURRENT_TIMESTAMP
	`, migration.Version, migration.Name, checksum, options.RuntimeVersion, executionTime, "applied")

	if err != nil {
		return nil, fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit migration: %w", err)
	}

	return &MigrationResult{
		Version:         migration.Version,
		Name:            migration.Name,
		Status:          "applied",
		ExecutionTimeMs: executionTime,
	}, nil
}

// rollbackMigrationV2 rolls back a single migration
func rollbackMigrationV2(db *sql.DB, migration *MigrationMetadata, options *MigrationOptions) (*MigrationResult, error) {
	startTime := time.Now()

	// Get migration file paths
	_, downPath := GetMigrationPath("migrations_v2", migration.Version)
	if downPath == "" {
		return nil, fmt.Errorf("rollback file not found for migration %s", migration.Version)
	}

	// Check if down.sql exists
	if _, err := os.Stat(downPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("rollback file does not exist for migration %s", migration.Version)
	}

	// Read rollback file
	downSQL, err := os.ReadFile(downPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rollback file: %w", err)
	}

	if options.DryRun {
		log.Printf("DRY RUN: Would rollback migration %s (%s)", migration.Version, migration.Name)
		return &MigrationResult{
			Version: migration.Version,
			Name:    migration.Name,
			Status:  "dry_run",
		}, nil
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Execute rollback
	if _, err := tx.Exec(string(downSQL)); err != nil {
		return nil, fmt.Errorf("failed to execute rollback: %w", err)
	}

	// Remove migration record
	_, err = tx.Exec("DELETE FROM schema_migrations_v2 WHERE version = $1", migration.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit rollback: %w", err)
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	return &MigrationResult{
		Version:         migration.Version,
		Name:            migration.Name,
		Status:          "rolled_back",
		ExecutionTimeMs: executionTime,
	}, nil
}

// calculateChecksum calculates SHA256 checksum of migration content
func calculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// ValidateMigrationIntegrity checks if migration files match their recorded checksums
func ValidateMigrationIntegrity(db *sql.DB) error {
	// Load migration registry
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return fmt.Errorf("failed to load migration registry: %w", err)
	}

	// Get applied migrations (ignore error if table doesn't exist yet)
	appliedMigrations, err := getAppliedMigrationsV2(db)
	if err != nil {
		// If table doesn't exist, that's fine - no migrations applied yet
		if err.Error() == `pq: relation "schema_migrations_v2" does not exist` {
			appliedMigrations = []MigrationStatus{}
		} else {
			return fmt.Errorf("failed to get applied migrations: %w", err)
		}
	}

	var issues []string

	for _, migration := range appliedMigrations {
		_, exists := registry.GetMigration(migration.Version)
		if !exists {
			issues = append(issues, fmt.Sprintf("Migration %s metadata not found", migration.Version))
			continue
		}

		// Get migration file path
		upPath, _ := GetMigrationPath("migrations_v2", migration.Version)
		if upPath == "" {
			issues = append(issues, fmt.Sprintf("Migration file not found for %s", migration.Version))
			continue
		}

		// Read current file content
		content, err := os.ReadFile(upPath)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Failed to read migration file %s: %v", migration.Version, err))
			continue
		}

		// Calculate current checksum
		currentChecksum := calculateChecksum(content)

		// Compare with recorded checksum
		if migration.Checksum != "" && migration.Checksum != currentChecksum {
			issues = append(issues, fmt.Sprintf("Migration %s checksum mismatch (recorded: %s, current: %s)", 
				migration.Version, migration.Checksum, currentChecksum))
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("migration integrity issues found:\n%s", fmt.Sprintf("  • %s\n", issues))
	}

	log.Println("All migration files are valid")
	return nil
}

