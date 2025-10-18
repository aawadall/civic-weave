package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"civicweave/backend/config"
	"civicweave/backend/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Parse command line arguments
	var (
		command            = flag.String("command", "up", "Migration command: up, down, status, compatibility, validate, check")
		targetVersion      = flag.String("version", "", "Target version for rollback")
		runtimeVersion     = flag.String("runtime-version", "1.0.0", "Runtime version for compatibility checking")
		dryRun             = flag.Bool("dry-run", false, "Show what would be done without executing")
		failOnIncompatible = flag.Bool("fail-on-incompatible", false, "Fail if runtime version is incompatible")
		quiet              = flag.Bool("quiet", false, "Suppress output (useful for CI/CD)")
	)
	flag.Parse()

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
