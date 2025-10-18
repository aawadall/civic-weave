package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// CompatibilityMatrix represents the compatibility status between database and runtime
type CompatibilityMatrix struct {
	CurrentDBVersion    string             `json:"current_db_version"`
	RuntimeVersion      string             `json:"runtime_version"`
	OverallStatus       string             `json:"overall_status"` // compatible, warning, incompatible
	AppliedMigrations   []MigrationStatus  `json:"applied_migrations"`
	PendingMigrations   []MigrationStatus  `json:"pending_migrations"`
	CompatibilityIssues []string           `json:"compatibility_issues,omitempty"`
}

// CheckDatabaseCompatibility checks the overall compatibility between database and runtime
func CheckDatabaseCompatibility(db *sql.DB, runtimeVersion string) (*CompatibilityMatrix, error) {
	matrix := &CompatibilityMatrix{
		RuntimeVersion: runtimeVersion,
	}

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

	// Get current database version (highest applied migration)
	if len(appliedMigrations) > 0 {
		matrix.CurrentDBVersion = appliedMigrations[len(appliedMigrations)-1].Version
	}

	// Load migration registry
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return nil, fmt.Errorf("failed to load migration registry: %w", err)
	}

	// Check compatibility for each applied migration
	allCompatible := true
	hasWarnings := false

	for _, migration := range appliedMigrations {
		compat, err := registry.CheckCompatibility(migration.Version, runtimeVersion)
		if err != nil {
			log.Printf("Warning: failed to check compatibility for migration %s: %v", migration.Version, err)
			continue
		}

		if !compat.IsCompatible {
			allCompatible = false
			matrix.CompatibilityIssues = append(matrix.CompatibilityIssues, 
				fmt.Sprintf("Migration %s: %s", migration.Version, compat.Message))
		} else if compat.Status == "warning" {
			hasWarnings = true
		}
	}

	// Determine overall status
	if !allCompatible {
		matrix.OverallStatus = "incompatible"
	} else if hasWarnings {
		matrix.OverallStatus = "warning"
	} else {
		matrix.OverallStatus = "compatible"
	}

	matrix.AppliedMigrations = appliedMigrations

	// Get pending migrations
	pendingMigrations, err := getPendingMigrations(db, registry)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending migrations: %w", err)
	}
	matrix.PendingMigrations = pendingMigrations

	return matrix, nil
}

// getAppliedMigrationsV2 retrieves applied migrations from schema_migrations_v2 table
func getAppliedMigrationsV2(db *sql.DB) ([]MigrationStatus, error) {
	query := `
		SELECT version, name, status, applied_at, runtime_version, execution_time_ms, checksum
		FROM schema_migrations_v2
		ORDER BY applied_at ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []MigrationStatus
	for rows.Next() {
		var migration MigrationStatus
		var appliedAt sql.NullString
		var runtimeVersion sql.NullString
		var executionTime sql.NullInt32
		var checksum sql.NullString

		err := rows.Scan(
			&migration.Version,
			&migration.Name,
			&migration.Status,
			&appliedAt,
			&runtimeVersion,
			&executionTime,
			&checksum,
		)
		if err != nil {
			return nil, err
		}

		if appliedAt.Valid {
			migration.AppliedAt = appliedAt.String
		}
		if runtimeVersion.Valid {
			migration.RuntimeVersion = runtimeVersion.String
		}
		if executionTime.Valid {
			migration.ExecutionTimeMs = int(executionTime.Int32)
		}
		if checksum.Valid {
			migration.Checksum = checksum.String
		}

		migrations = append(migrations, migration)
	}

	return migrations, rows.Err()
}

// getPendingMigrations returns migrations that haven't been applied yet
func getPendingMigrations(db *sql.DB, registry *MigrationRegistry) ([]MigrationStatus, error) {
	// Get all available migrations
	allMigrations := registry.GetSortedMigrations()

	// Get applied migration versions (ignore error if table doesn't exist yet)
	appliedVersions := make(map[string]bool)
	appliedMigrations, err := getAppliedMigrationsV2(db)
	if err != nil {
		// If table doesn't exist, that's fine - no migrations applied yet
		if err.Error() == `pq: relation "schema_migrations_v2" does not exist` {
			appliedMigrations = []MigrationStatus{}
		} else {
			return nil, err
		}
	}

	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	// Find pending migrations
	var pending []MigrationStatus
	for _, migration := range allMigrations {
		if !appliedVersions[migration.Version] {
			pending = append(pending, MigrationStatus{
				Version: migration.Version,
				Name:    migration.Name,
				Status:  "pending",
			})
		}
	}

	return pending, nil
}

// DisplayCompatibilityMatrix prints a formatted compatibility matrix
func DisplayCompatibilityMatrix(matrix *CompatibilityMatrix) {
	fmt.Println("🔍 Database Compatibility Matrix")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Overall status
	statusIcon := "✅"
	if matrix.OverallStatus == "warning" {
		statusIcon = "⚠️"
	} else if matrix.OverallStatus == "incompatible" {
		statusIcon = "❌"
	}

	fmt.Printf("%s Overall Status: %s\n", statusIcon, strings.ToUpper(matrix.OverallStatus))
	fmt.Printf("📊 Database Version: %s\n", matrix.CurrentDBVersion)
	fmt.Printf("🚀 Runtime Version: %s\n", matrix.RuntimeVersion)
	fmt.Println()

	// Applied migrations
	if len(matrix.AppliedMigrations) > 0 {
		fmt.Println("📋 Applied Migrations:")
		for _, migration := range matrix.AppliedMigrations {
			statusIcon := "✅"
			if migration.Status == "failed" {
				statusIcon = "❌"
			}
			fmt.Printf("  %s %s (%s) - %s\n", 
				statusIcon, migration.Version, migration.Name, migration.AppliedAt)
		}
		fmt.Println()
	}

	// Pending migrations
	if len(matrix.PendingMigrations) > 0 {
		fmt.Println("⏳ Pending Migrations:")
		for _, migration := range matrix.PendingMigrations {
			fmt.Printf("  ⏸️  %s (%s)\n", migration.Version, migration.Name)
		}
		fmt.Println()
	}

	// Compatibility issues
	if len(matrix.CompatibilityIssues) > 0 {
		fmt.Println("⚠️  Compatibility Issues:")
		for _, issue := range matrix.CompatibilityIssues {
			fmt.Printf("  • %s\n", issue)
		}
		fmt.Println()
	}

	// Recommendations
	fmt.Println("💡 Recommendations:")
	if matrix.OverallStatus == "incompatible" {
		fmt.Println("  • Update runtime version to meet minimum requirements")
		fmt.Println("  • Or rollback database to a compatible version")
	} else if matrix.OverallStatus == "warning" {
		fmt.Println("  • Consider updating database migrations")
		fmt.Println("  • Monitor for any compatibility issues")
	} else if len(matrix.PendingMigrations) > 0 {
		fmt.Println("  • Run migrations to update database schema")
		fmt.Println("  • Use 'make db-migrate-v2' to apply pending migrations")
	} else {
		fmt.Println("  • Database is up to date and compatible")
	}
}

// GetMinimumRequiredVersion returns the minimum runtime version required for current DB state
func GetMinimumRequiredVersion(db *sql.DB) (string, error) {
	appliedMigrations, err := getAppliedMigrationsV2(db)
	if err != nil {
		// If table doesn't exist, that's fine - no migrations applied yet
		if err.Error() == `pq: relation "schema_migrations_v2" does not exist` {
			return "0.0.0", nil
		}
		return "", err
	}

	if len(appliedMigrations) == 0 {
		return "0.0.0", nil // No migrations applied, any version is fine
	}

	// Load migration registry
	registry, err := LoadMigrationsFromDirectory("migrations_v2")
	if err != nil {
		return "", err
	}

	// Find the highest minimum runtime version requirement
	var maxMinVersion *semver.Version
	for _, migration := range appliedMigrations {
		metadata, exists := registry.GetMigration(migration.Version)
		if !exists {
			continue
		}

		if metadata.MinRuntimeVersion != "" {
			minVer, err := semver.NewVersion(metadata.MinRuntimeVersion)
			if err != nil {
				continue
			}

			if maxMinVersion == nil || minVer.GreaterThan(maxMinVersion) {
				maxMinVersion = minVer
			}
		}
	}

	if maxMinVersion == nil {
		return "0.0.0", nil
	}

	return maxMinVersion.String(), nil
}

// ValidateRuntimeVersion checks if runtime version meets minimum requirements
func ValidateRuntimeVersion(db *sql.DB, runtimeVersion string) error {
	minVersion, err := GetMinimumRequiredVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get minimum required version: %w", err)
	}

	if minVersion == "0.0.0" {
		return nil // No requirements
	}

	runtimeVer, err := semver.NewVersion(runtimeVersion)
	if err != nil {
		return fmt.Errorf("invalid runtime version: %w", err)
	}

	minVer, err := semver.NewVersion(minVersion)
	if err != nil {
		return fmt.Errorf("invalid minimum version: %w", err)
	}

	if runtimeVer.LessThan(minVer) {
		return fmt.Errorf("runtime version %s is below minimum required %s", runtimeVersion, minVersion)
	}

	return nil
}
