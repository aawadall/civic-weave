package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"civicweave/backend/config"
	"civicweave/backend/database"

	"github.com/joho/godotenv"
)

func main() {
	// Parse command line arguments
	var (
		command            = flag.String("command", "up", "Migration command: up, down, status, compatibility, validate, check, schema-state, drift-detect, validate-state")
		targetVersion      = flag.String("version", "", "Target version for rollback")
		runtimeVersion     = flag.String("runtime-version", "1.0.0", "Runtime version for compatibility checking")
		dryRun             = flag.Bool("dry-run", false, "Show what would be done without executing")
		failOnIncompatible = flag.Bool("fail-on-incompatible", false, "Fail if runtime version is incompatible")
		quiet              = flag.Bool("quiet", false, "Suppress output (useful for CI/CD)")
		envFile            = flag.String("env", ".env", "Environment file path")
	)
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("No %s file found, using system environment variables", *envFile)
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Execute command
	switch *command {
	case "up":
		err = runMigrations(db, *runtimeVersion, *dryRun, *failOnIncompatible, *quiet)
	case "down":
		if *targetVersion == "" {
			log.Fatal("Target version is required for rollback")
		}
		err = rollbackMigrations(db, *targetVersion, *dryRun, *quiet)
	case "status":
		err = showMigrationStatus(db, *quiet)
	case "compatibility":
		err = showCompatibilityMatrix(db, *runtimeVersion, *quiet)
	case "validate":
		err = validateMigrations(db, *quiet)
	case "check":
		exitCode, message, err := checkMigrationHealth(db, *runtimeVersion, *quiet)
		if err != nil {
			log.Fatal(err)
		}
		if !*quiet {
			fmt.Println(message)
		}
		os.Exit(exitCode)
	case "schema-state":
		err = showSchemaState(db, *quiet)
	case "drift-detect":
		err = detectSchemaDrift(db, *quiet)
	case "validate-state":
		if *targetVersion == "" {
			log.Fatal("Target version is required for state validation")
		}
		err = validateIntendedState(db, *targetVersion, *quiet)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}

	if err != nil {
		log.Fatal(err)
	}

	if !*quiet {
		log.Println("Operation completed successfully")
	}
}

func runMigrations(db *sql.DB, runtimeVersion string, dryRun, failOnIncompatible, quiet bool) error {
	options := &database.MigrationHookOptions{
		DryRun:             dryRun,
		FailOnIncompatible: failOnIncompatible,
		RuntimeVersion:     runtimeVersion,
		Quiet:              quiet,
	}

	return database.AutoMigrate(db, runtimeVersion, options)
}

func rollbackMigrations(db *sql.DB, targetVersion string, dryRun, quiet bool) error {
	options := &database.MigrationHookOptions{
		DryRun: dryRun,
		Quiet:  quiet,
	}

	return database.RollbackToVersion(db, targetVersion, options)
}

func showMigrationStatus(db *sql.DB, quiet bool) error {
	status, err := database.GetMigrationStatus(db)
	if err != nil {
		return err
	}

	if quiet {
		// JSON output for programmatic use
		fmt.Printf(`{"current_db_version":"%s","applied_count":%d,"pending_count":%d,"overall_status":"%s"}`,
			status.CurrentDBVersion, len(status.AppliedMigrations), len(status.PendingMigrations), status.OverallStatus)
		return nil
	}

	// Human-readable output
	fmt.Println("📊 Migration Status")
	fmt.Println("=" + fmt.Sprintf("%*s", 30, "="))

	if status.CurrentDBVersion != "" {
		fmt.Printf("Current DB Version: %s\n", status.CurrentDBVersion)
	} else {
		fmt.Println("Current DB Version: None")
	}

	fmt.Printf("Applied Migrations: %d\n", len(status.AppliedMigrations))
	fmt.Printf("Pending Migrations: %d\n", len(status.PendingMigrations))
	fmt.Printf("Overall Status: %s\n", status.OverallStatus)

	if len(status.AppliedMigrations) > 0 {
		fmt.Println("\n📋 Applied Migrations:")
		for _, migration := range status.AppliedMigrations {
			fmt.Printf("  ✅ %s (%s) - %s\n", migration.Version, migration.Name, migration.AppliedAt)
		}
	}

	if len(status.PendingMigrations) > 0 {
		fmt.Println("\n⏳ Pending Migrations:")
		for _, migration := range status.PendingMigrations {
			fmt.Printf("  ⏸️  %s (%s)\n", migration.Version, migration.Name)
		}
	}

	return nil
}

func showCompatibilityMatrix(db *sql.DB, runtimeVersion string, quiet bool) error {
	matrix, err := database.CheckDatabaseCompatibility(db, runtimeVersion)
	if err != nil {
		return err
	}

	if quiet {
		// JSON output for programmatic use
		fmt.Printf(`{"overall_status":"%s","current_db_version":"%s","runtime_version":"%s","is_compatible":%t}`,
			matrix.OverallStatus, matrix.CurrentDBVersion, matrix.RuntimeVersion, matrix.OverallStatus == "compatible")
		return nil
	}

	database.DisplayCompatibilityMatrix(matrix)
	return nil
}

func validateMigrations(db *sql.DB, quiet bool) error {
	// Validate migration files
	if err := database.ValidateMigrationFiles(); err != nil {
		return fmt.Errorf("migration files validation failed: %w", err)
	}

	// Validate migration integrity
	if err := database.ValidateMigrationIntegrity(db); err != nil {
		return fmt.Errorf("migration integrity validation failed: %w", err)
	}

	if !quiet {
		fmt.Println("✅ All migration files are valid")
		fmt.Println("✅ Migration integrity check passed")
	}

	return nil
}

func checkMigrationHealth(db *sql.DB, runtimeVersion string, quiet bool) (int, string, error) {
	return database.CheckMigrationHealth(db, runtimeVersion)
}

func showSchemaState(db *sql.DB, quiet bool) error {
	state, err := database.GetCurrentSchemaState(db)
	if err != nil {
		return err
	}

	if quiet {
		// JSON output for programmatic use
		fmt.Printf(`{"checksum":"%s","tables":%d,"indexes":%d,"functions":%d}`,
			state.Checksum, len(state.Tables), len(state.Indexes), len(state.Functions))
		return nil
	}

	// Human-readable output
	fmt.Println("📊 Database Schema State")
	fmt.Println("=" + fmt.Sprintf("%*s", 30, "="))
	fmt.Printf("Schema Checksum: %s\n", state.Checksum)
	fmt.Printf("Tables: %d\n", len(state.Tables))
	fmt.Printf("Indexes: %d\n", len(state.Indexes))
	fmt.Printf("Functions: %d\n", len(state.Functions))

	if len(state.Tables) > 0 {
		fmt.Println("\n📋 Tables:")
		for _, table := range state.Tables {
			fmt.Printf("  📄 %s (%d columns) - %s\n", table.Name, len(table.Columns), table.Checksum[:8])
		}
	}

	if len(state.Indexes) > 0 {
		fmt.Println("\n🔍 Indexes:")
		for _, index := range state.Indexes {
			unique := ""
			if index.IsUnique {
				unique = " (unique)"
			}
			fmt.Printf("  📇 %s.%s%s - %s\n", index.TableName, index.Name, unique, strings.Join(index.Columns, ", "))
		}
	}

	if len(state.Functions) > 0 {
		fmt.Println("\n⚙️  Functions:")
		for _, function := range state.Functions {
			fmt.Printf("  🔧 %s - %s\n", function.Name, function.Checksum[:8])
		}
	}

	return nil
}

func detectSchemaDrift(db *sql.DB, quiet bool) error {
	comparison, err := database.DetectSchemaDrift(db)
	if err != nil {
		return err
	}

	if quiet {
		// JSON output for programmatic use
		fmt.Printf(`{"is_identical":%t,"checksum_match":%t,"drift_detected":%t}`,
			comparison.IsIdentical, comparison.ChecksumMatch, !comparison.IsIdentical)
		return nil
	}

	// Human-readable output
	fmt.Println("🔍 Schema Drift Detection")
	fmt.Println("=" + fmt.Sprintf("%*s", 30, "="))

	if comparison.IsIdentical {
		fmt.Println("✅ No schema drift detected")
		fmt.Println("📊 Schema is consistent with expected state")
	} else {
		fmt.Println("⚠️  Schema drift detected")
		fmt.Printf("📊 Local checksum:  %s\n", comparison.LocalChecksum)
		fmt.Printf("📊 Remote checksum: %s\n", comparison.RemoteChecksum)
	}

	if len(comparison.Differences) > 0 {
		fmt.Println("\n📋 Differences:")
		for _, diff := range comparison.Differences {
			fmt.Printf("  • %s\n", diff)
		}
	}

	if len(comparison.MissingTables) > 0 {
		fmt.Println("\n❌ Missing Tables:")
		for _, table := range comparison.MissingTables {
			fmt.Printf("  • %s\n", table)
		}
	}

	if len(comparison.ExtraTables) > 0 {
		fmt.Println("\n➕ Extra Tables:")
		for _, table := range comparison.ExtraTables {
			fmt.Printf("  • %s\n", table)
		}
	}

	if len(comparison.SchemaDrift) > 0 {
		fmt.Println("\n🔄 Schema Drift:")
		for _, drift := range comparison.SchemaDrift {
			fmt.Printf("  • %s\n", drift)
		}
	}

	return nil
}

func validateIntendedState(db *sql.DB, targetVersion string, quiet bool) error {
	err := database.ValidateIntendedState(db, targetVersion)
	if err != nil {
		return err
	}

	if !quiet {
		fmt.Printf("✅ Database state validated for version %s\n", targetVersion)
		fmt.Println("📊 Database matches intended state")
	}

	return nil
}
