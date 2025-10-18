package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
)

// MigrationHookOptions represents options for programmatic migration operations
type MigrationHookOptions struct {
	DryRun             bool   `json:"dry_run"`
	FailOnIncompatible bool   `json:"fail_on_incompatible"`
	MaxVersion         string `json:"max_version"`
	AutoApprove        bool   `json:"auto_approve"`
	RuntimeVersion     string `json:"runtime_version"`
	Quiet              bool   `json:"quiet"`
}

// CheckCompatibility returns compatibility status without printing
func CheckCompatibility(db *sql.DB, runtimeVersion string) (*CompatibilityStatus, error) {
	// Get minimum required version
	minVersion, err := GetMinimumRequiredVersion(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get minimum required version: %w", err)
	}

	if minVersion == "0.0.0" {
		return &CompatibilityStatus{
			IsCompatible:     true,
			Status:           "compatible",
			CurrentDBVersion: "none",
			RequiredVersion:  "0.0.0",
			Message:          "No migrations applied, any version is compatible",
		}, nil
	}

	// Parse versions
	runtimeVer, err := parseVersion(runtimeVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid runtime version: %w", err)
	}

	minVer, err := parseVersion(minVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid minimum version: %w", err)
	}

	// Check compatibility
	if runtimeVer.LessThan(minVer) {
		return &CompatibilityStatus{
			IsCompatible:     false,
			Status:           "incompatible",
			CurrentDBVersion: minVersion,
			RequiredVersion:  minVersion,
			Message:          fmt.Sprintf("Runtime version %s is below minimum required %s", runtimeVersion, minVersion),
		}, nil
	}

	return &CompatibilityStatus{
		IsCompatible:     true,
		Status:           "compatible",
		CurrentDBVersion: minVersion,
		RequiredVersion:  minVersion,
		Message:          "Runtime version is compatible with database",
	}, nil
}

// GetMigrationStatus returns structured migration status data
func GetMigrationStatus(db *sql.DB) (*CompatibilityMatrix, error) {
	// Get applied migrations (ignore error if table doesn't exist yet)
	appliedMigrations, err := getAppliedMigrationsV2(db)
	if err != nil {
		// If table doesn't exist, that's fine - no migrations applied yet
		if err.Error() == `pq: relation "schema_migrations_v2" does not exist` {
			appliedMigrations = []MigrationStatus{}
		} else {
			return nil, fmt.Errorf("failed to get applied migrations: %w", err)
		}
	}

	// Load migration registry
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return nil, fmt.Errorf("failed to load migration registry: %w", err)
	}

	// Get current database version
	var currentDBVersion string
	if len(appliedMigrations) > 0 {
		currentDBVersion = appliedMigrations[len(appliedMigrations)-1].Version
	}

	// Get pending migrations
	pendingMigrations, err := getPendingMigrations(db, registry)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending migrations: %w", err)
	}

	status := &CompatibilityMatrix{
		CurrentDBVersion:  currentDBVersion,
		AppliedMigrations: appliedMigrations,
		PendingMigrations: pendingMigrations,
		OverallStatus:     "unknown",
	}

	// Determine overall status
	if len(pendingMigrations) > 0 {
		status.OverallStatus = "pending_migrations"
	} else if len(appliedMigrations) > 0 {
		status.OverallStatus = "up_to_date"
	} else {
		status.OverallStatus = "no_migrations"
	}

	return status, nil
}

// AutoMigrate performs safe auto-migration with configurable behavior
func AutoMigrate(db *sql.DB, runtimeVersion string, options *MigrationHookOptions) error {
	if options == nil {
		options = &MigrationHookOptions{
			RuntimeVersion: runtimeVersion,
		}
	}

	// Check compatibility first
	compat, err := CheckCompatibility(db, runtimeVersion)
	if err != nil {
		return fmt.Errorf("compatibility check failed: %w", err)
	}

	if !compat.IsCompatible {
		if options.FailOnIncompatible {
			return fmt.Errorf("runtime version %s is incompatible with database: %s", runtimeVersion, compat.Message)
		}
		if !options.Quiet {
			log.Printf("Warning: %s", compat.Message)
		}
	}

	// Run migrations
	migrateOptions := &MigrationOptions{
		DryRun:             options.DryRun,
		FailOnIncompatible: options.FailOnIncompatible,
		MaxVersion:         options.MaxVersion,
		AutoApprove:        options.AutoApprove,
		RuntimeVersion:     runtimeVersion,
	}
	return MigrateV2(db, migrateOptions)
}

// CheckMigrationHealth performs a health check for CI/CD pipelines
func CheckMigrationHealth(db *sql.DB, runtimeVersion string) (int, string, error) {
	// Check compatibility
	compat, err := CheckCompatibility(db, runtimeVersion)
	if err != nil {
		return 2, "", fmt.Errorf("compatibility check failed: %w", err)
	}

	if !compat.IsCompatible {
		return 2, compat.Message, nil // Exit code 2: incompatible
	}

	// Check for pending migrations
	status, err := GetMigrationStatus(db)
	if err != nil {
		return 2, "", fmt.Errorf("failed to get migration status: %w", err)
	}

	if len(status.PendingMigrations) > 0 {
		return 1, fmt.Sprintf("Pending migrations: %d", len(status.PendingMigrations)), nil // Exit code 1: needs migration
	}

	return 0, "Database is up to date and compatible", nil // Exit code 0: compatible
}

// ValidateMigrationFiles checks if all migration files are valid
func ValidateMigrationFiles() error {
	// Check if migrations_v2 directory exists
	if _, err := os.Stat("migrations_v2"); os.IsNotExist(err) {
		return fmt.Errorf("migrations_v2 directory not found")
	}

	// Load migration registry
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return fmt.Errorf("failed to load migration registry: %w", err)
	}

	// Validate each migration
	var issues []string
	for version, migration := range registry.Migrations {
		// Check if migration directory exists
		upPath, downPath := GetMigrationPath("migrations_v2", version)
		if upPath == "" {
			issues = append(issues, fmt.Sprintf("Migration %s: up.sql not found", version))
			continue
		}

		// Check up.sql exists and is readable
		if _, err := os.Stat(upPath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Migration %s: up.sql file not found", version))
		}

		// Check down.sql exists (optional but recommended)
		if _, err := os.Stat(downPath); os.IsNotExist(err) {
			if migration.Name != "initial_schema" { // Initial schema might not need rollback
				issues = append(issues, fmt.Sprintf("Migration %s: down.sql not found (rollback not possible)", version))
			}
		}

		// Validate metadata
		if migration.Version == "" {
			issues = append(issues, fmt.Sprintf("Migration %s: version is required", version))
		}
		if migration.Name == "" {
			issues = append(issues, fmt.Sprintf("Migration %s: name is required", version))
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("migration validation failed:\n%s", fmt.Sprintf("  • %s\n", issues))
	}

	return nil
}

// GetMigrationSummary returns a summary of migration status
func GetMigrationSummary(db *sql.DB) (map[string]interface{}, error) {
	status, err := GetMigrationStatus(db)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"current_db_version": status.CurrentDBVersion,
		"applied_count":      len(status.AppliedMigrations),
		"pending_count":      len(status.PendingMigrations),
		"overall_status":     status.OverallStatus,
	}

	// Add applied migrations list
	var appliedList []map[string]interface{}
	for _, migration := range status.AppliedMigrations {
		appliedList = append(appliedList, map[string]interface{}{
			"version":    migration.Version,
			"name":       migration.Name,
			"applied_at": migration.AppliedAt,
			"status":     migration.Status,
		})
	}
	summary["applied_migrations"] = appliedList

	// Add pending migrations list
	var pendingList []map[string]interface{}
	for _, migration := range status.PendingMigrations {
		pendingList = append(pendingList, map[string]interface{}{
			"version": migration.Version,
			"name":    migration.Name,
			"status":  migration.Status,
		})
	}
	summary["pending_migrations"] = pendingList

	return summary, nil
}

// RollbackToVersion rolls back to a specific version
func RollbackToVersion(db *sql.DB, targetVersion string, options *MigrationHookOptions) error {
	if options == nil {
		options = &MigrationHookOptions{}
	}

	rollbackOptions := &MigrationOptions{
		DryRun: options.DryRun,
	}
	return RollbackV2(db, targetVersion, rollbackOptions)
}

// GetAvailableMigrations returns all available migrations
func GetAvailableMigrations() ([]*MigrationMetadata, error) {
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return nil, err
	}

	return registry.GetSortedMigrations(), nil
}

// GetMigrationDependencies returns dependency information for a migration
func GetMigrationDependencies(version string) ([]string, error) {
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return nil, err
	}

	migration, exists := registry.GetMigration(version)
	if !exists {
		return nil, fmt.Errorf("migration %s not found", version)
	}

	return migration.Dependencies, nil
}

// CheckMigrationDependencies validates that all dependencies are satisfied
func CheckMigrationDependencies(db *sql.DB, version string) error {
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return err
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrationsV2(db)
	if err != nil {
		return err
	}

	appliedVersions := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	return registry.ValidateDependencies(version, appliedVersions)
}

// parseVersion is a helper function to parse semantic versions
func parseVersion(version string) (*semver.Version, error) {
	return semver.NewVersion(version)
}
