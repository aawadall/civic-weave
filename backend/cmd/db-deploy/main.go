package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"civicweave/backend/config"
	"civicweave/backend/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	var (
		envFile    = flag.String("env", ".env", "Environment file path")
		dryRun     = flag.Bool("dry-run", false, "Show what would be executed without running")
		version    = flag.String("version", "", "Run migration up to specific version (e.g., 011)")
		rollback   = flag.String("rollback", "", "Rollback to specific version (e.g., 010)")
		status     = flag.Bool("status", false, "Show migration status")
		help       = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

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

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Printf("✅ Connected to database: %s", cfg.Database.Host)

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatal("Failed to create migrations table:", err)
	}

	if *status {
		showMigrationStatus(db)
		return
	}

	if *rollback != "" {
		if err := rollbackMigration(db, *rollback, *dryRun); err != nil {
			log.Fatal("Failed to rollback migration:", err)
		}
		return
	}

	// Run migrations
	if err := runMigrations(db, *version, *dryRun); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
}

func showHelp() {
	fmt.Println("Database Migration Utility")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/db-deploy/main.go [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -env string")
	fmt.Println("        Environment file path (default \".env\")")
	fmt.Println("  -dry-run")
	fmt.Println("        Show what would be executed without running")
	fmt.Println("  -version string")
	fmt.Println("        Run migration up to specific version (e.g., \"011\")")
	fmt.Println("  -rollback string")
	fmt.Println("        Rollback to specific version (e.g., \"010\")")
	fmt.Println("  -status")
	fmt.Println("        Show migration status")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Show migration status")
	fmt.Println("  go run cmd/db-deploy/main.go -status")
	fmt.Println("")
	fmt.Println("  # Run all pending migrations")
	fmt.Println("  go run cmd/db-deploy/main.go")
	fmt.Println("")
	fmt.Println("  # Run migrations up to version 011")
	fmt.Println("  go run cmd/db-deploy/main.go -version 011")
	fmt.Println("")
	fmt.Println("  # Dry run to see what would be executed")
	fmt.Println("  go run cmd/db-deploy/main.go -dry-run")
	fmt.Println("")
	fmt.Println("  # Rollback to version 010")
	fmt.Println("  go run cmd/db-deploy/main.go -rollback 010")
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		checksum VARCHAR(255)
	)`
	
	_, err := db.Exec(query)
	return err
}

func showMigrationStatus(db *sql.DB) {
	fmt.Println("📊 Migration Status")
	fmt.Println("==================")
	
	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		log.Fatal("Failed to get applied migrations:", err)
	}
	
	// Get available migrations
	available, err := getAvailableMigrations()
	if err != nil {
		log.Fatal("Failed to get available migrations:", err)
	}
	
	fmt.Printf("Applied migrations: %d\n", len(applied))
	fmt.Printf("Available migrations: %d\n", len(available))
	fmt.Println("")
	
	// Show status for each migration
	for _, migration := range available {
		status := "❌ Pending"
		if _, exists := applied[migration.Version]; exists {
			status = "✅ Applied"
		}
		fmt.Printf("%s %s - %s\n", status, migration.Version, migration.Description)
	}
	
	// Show pending migrations
	pending := getPendingMigrations(applied, available)
	if len(pending) > 0 {
		fmt.Printf("\n🚀 %d pending migrations ready to apply:\n", len(pending))
		for _, migration := range pending {
			fmt.Printf("  - %s: %s\n", migration.Version, migration.Description)
		}
	} else {
		fmt.Println("\n✅ All migrations are up to date!")
	}
}

type Migration struct {
	Version     string
	Description string
	Path        string
}

func getAvailableMigrations() ([]Migration, error) {
	migrationsDir := "migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return nil, err
	}
	
	var migrations []Migration
	for _, file := range files {
		filename := filepath.Base(file)
		// Extract version from filename like "011_task_enhancements.sql"
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			continue
		}
		
		version := parts[0]
		description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")
		description = strings.ReplaceAll(description, "_", " ")
		
		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			Path:        file,
		})
	}
	
	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	
	return migrations, nil
}

func getAppliedMigrations(db *sql.DB) (map[string]string, error) {
	query := "SELECT version, checksum FROM schema_migrations ORDER BY version"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	applied := make(map[string]string)
	for rows.Next() {
		var version, checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, err
		}
		applied[version] = checksum
	}
	
	return applied, rows.Err()
}

func getPendingMigrations(applied map[string]string, available []Migration) []Migration {
	var pending []Migration
	for _, migration := range available {
		if _, exists := applied[migration.Version]; !exists {
			pending = append(pending, migration)
		}
	}
	return pending
}

func runMigrations(db *sql.DB, targetVersion string, dryRun bool) error {
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}
	
	available, err := getAvailableMigrations()
	if err != nil {
		return err
	}
	
	pending := getPendingMigrations(applied, available)
	
	if len(pending) == 0 {
		fmt.Println("✅ All migrations are up to date!")
		return nil
	}
	
	// Filter by target version if specified
	if targetVersion != "" {
		var filtered []Migration
		for _, migration := range pending {
			if migration.Version <= targetVersion {
				filtered = append(filtered, migration)
			}
		}
		pending = filtered
	}
	
	if len(pending) == 0 {
		fmt.Printf("✅ No migrations to run up to version %s\n", targetVersion)
		return nil
	}
	
	fmt.Printf("🚀 Running %d migrations...\n", len(pending))
	
	for _, migration := range pending {
		if err := runMigration(db, migration, dryRun); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
		}
	}
	
	fmt.Println("✅ All migrations completed successfully!")
	return nil
}

func runMigration(db *sql.DB, migration Migration, dryRun bool) error {
	fmt.Printf("📝 Running migration %s: %s\n", migration.Version, migration.Description)
	
	// Read migration file
	content, err := os.ReadFile(migration.Path)
	if err != nil {
		return err
	}
	
	// Split into UP and DOWN sections
	sections := strings.Split(string(content), "-- DOWN")
	if len(sections) < 2 {
		return fmt.Errorf("migration file must contain -- UP and -- DOWN sections")
	}
	
	upSection := strings.TrimSpace(sections[0])
	// Remove "-- UP" header
	if strings.HasPrefix(upSection, "-- UP") {
		upSection = strings.TrimSpace(strings.TrimPrefix(upSection, "-- UP"))
	}
	
	if dryRun {
		fmt.Printf("🔍 [DRY RUN] Would execute:\n%s\n", upSection)
		return nil
	}
	
	// Execute migration
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Split by semicolon and execute each statement
	statements := strings.Split(upSection, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\nStatement: %s", err, stmt)
		}
	}
	
	// Record migration as applied
	checksum := fmt.Sprintf("%x", len(content)) // Simple checksum
	_, err = tx.Exec("INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)", migration.Version, checksum)
	if err != nil {
		return err
	}
	
	if err := tx.Commit(); err != nil {
		return err
	}
	
	fmt.Printf("✅ Migration %s completed successfully\n", migration.Version)
	return nil
}

func rollbackMigration(db *sql.DB, targetVersion string, dryRun bool) error {
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}
	
	available, err := getAvailableMigrations()
	if err != nil {
		return err
	}
	
	// Find migrations to rollback (those with version > targetVersion)
	var toRollback []Migration
	for _, migration := range available {
		if _, exists := applied[migration.Version]; exists && migration.Version > targetVersion {
			toRollback = append(toRollback, migration)
		}
	}
	
	if len(toRollback) == 0 {
		fmt.Printf("✅ No migrations to rollback to version %s\n", targetVersion)
		return nil
	}
	
	// Sort in reverse order for rollback
	sort.Slice(toRollback, func(i, j int) bool {
		return toRollback[i].Version > toRollback[j].Version
	})
	
	fmt.Printf("🔄 Rolling back %d migrations to version %s...\n", len(toRollback), targetVersion)
	
	for _, migration := range toRollback {
		if err := rollbackSingleMigration(db, migration, dryRun); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
		}
	}
	
	fmt.Println("✅ Rollback completed successfully!")
	return nil
}

func rollbackSingleMigration(db *sql.DB, migration Migration, dryRun bool) error {
	fmt.Printf("🔄 Rolling back migration %s: %s\n", migration.Version, migration.Description)
	
	// Read migration file
	content, err := os.ReadFile(migration.Path)
	if err != nil {
		return err
	}
	
	// Split into UP and DOWN sections
	sections := strings.Split(string(content), "-- DOWN")
	if len(sections) < 2 {
		return fmt.Errorf("migration file must contain -- UP and -- DOWN sections")
	}
	
	downSection := strings.TrimSpace(sections[1])
	
	if dryRun {
		fmt.Printf("🔍 [DRY RUN] Would execute rollback:\n%s\n", downSection)
		return nil
	}
	
	// Execute rollback
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Split by semicolon and execute each statement
	statements := strings.Split(downSection, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute rollback statement: %w\nStatement: %s", err, stmt)
		}
	}
	
	// Remove migration from applied list
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = $1", migration.Version)
	if err != nil {
		return err
	}
	
	if err := tx.Commit(); err != nil {
		return err
	}
	
	fmt.Printf("✅ Migration %s rolled back successfully\n", migration.Version)
	return nil
}
