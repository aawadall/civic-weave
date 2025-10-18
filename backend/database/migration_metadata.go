package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// MigrationMetadata represents the metadata for a migration
type MigrationMetadata struct {
	Version            string   `json:"version"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	MinRuntimeVersion  string   `json:"min_runtime_version"`
	MaxRuntimeVersion  string   `json:"max_runtime_version,omitempty"`
	Dependencies       []string `json:"dependencies,omitempty"`
	Checksum           string   `json:"checksum,omitempty"`
	Author             string   `json:"author,omitempty"`
	CreatedAt          string   `json:"created_at,omitempty"`
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version         string `json:"version"`
	Name            string `json:"name"`
	Status          string `json:"status"` // applied, pending, failed
	AppliedAt       string `json:"applied_at,omitempty"`
	RuntimeVersion  string `json:"runtime_version,omitempty"`
	ExecutionTimeMs int    `json:"execution_time_ms,omitempty"`
	Checksum        string `json:"checksum,omitempty"`
}

// CompatibilityStatus represents the compatibility status between DB and runtime
type CompatibilityStatus struct {
	IsCompatible     bool   `json:"is_compatible"`
	Status           string `json:"status"` // compatible, warning, incompatible
	CurrentDBVersion string `json:"current_db_version"`
	RequiredVersion  string `json:"required_version"`
	Message          string `json:"message"`
}

// MigrationRegistry manages migration metadata
type MigrationRegistry struct {
	Migrations map[string]*MigrationMetadata
}

// NewMigrationRegistry creates a new migration registry
func NewMigrationRegistry() *MigrationRegistry {
	return &MigrationRegistry{
		Migrations: make(map[string]*MigrationMetadata),
	}
}

// LoadMigrationMetadata loads metadata from a JSON file
func LoadMigrationMetadata(filePath string) (*MigrationMetadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata MigrationMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	// Validate required fields
	if metadata.Version == "" {
		return nil, fmt.Errorf("version is required")
	}
	if metadata.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Validate version format (semantic versioning)
	if _, err := semver.NewVersion(metadata.Version); err != nil {
		return nil, fmt.Errorf("invalid semantic version: %s", metadata.Version)
	}

	return &metadata, nil
}

// LoadMigrationsFromDirectory loads all migrations from a directory
func LoadMigrationsFromDirectory(dirPath string) (*MigrationRegistry, error) {
	registry := NewMigrationRegistry()

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return registry, nil // Return empty registry if directory doesn't exist
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		migrationDir := filepath.Join(dirPath, entry.Name())
		metadataFile := filepath.Join(migrationDir, "metadata.json")

		// Check if metadata.json exists
		if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
			continue // Skip directories without metadata.json
		}

		metadata, err := LoadMigrationMetadata(metadataFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load migration %s: %w", entry.Name(), err)
		}

		registry.Migrations[metadata.Version] = metadata
	}

	return registry, nil
}

// GetSortedMigrations returns migrations sorted by version
func (r *MigrationRegistry) GetSortedMigrations() []*MigrationMetadata {
	var migrations []*MigrationMetadata
	for _, migration := range r.Migrations {
		migrations = append(migrations, migration)
	}

	// Sort by semantic version
	sort.Slice(migrations, func(i, j int) bool {
		v1, _ := semver.NewVersion(migrations[i].Version)
		v2, _ := semver.NewVersion(migrations[j].Version)
		return v1.LessThan(v2)
	})

	return migrations
}

// GetMigration returns a migration by version
func (r *MigrationRegistry) GetMigration(version string) (*MigrationMetadata, bool) {
	migration, exists := r.Migrations[version]
	return migration, exists
}

// ValidateDependencies checks if all dependencies are satisfied
func (r *MigrationRegistry) ValidateDependencies(version string, appliedMigrations map[string]bool) error {
	migration, exists := r.GetMigration(version)
	if !exists {
		return fmt.Errorf("migration %s not found", version)
	}

	for _, dep := range migration.Dependencies {
		if !appliedMigrations[dep] {
			return fmt.Errorf("dependency %s not satisfied for migration %s", dep, version)
		}
	}

	return nil
}

// CheckCompatibility checks if a migration is compatible with the runtime version
func (r *MigrationRegistry) CheckCompatibility(version, runtimeVersion string) (*CompatibilityStatus, error) {
	migration, exists := r.GetMigration(version)
	if !exists {
		return nil, fmt.Errorf("migration %s not found", version)
	}

	runtimeVer, err := semver.NewVersion(runtimeVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid runtime version: %w", err)
	}

	status := &CompatibilityStatus{
		CurrentDBVersion: version,
		RequiredVersion:  migration.MinRuntimeVersion,
	}

	// Check minimum runtime version
	if migration.MinRuntimeVersion != "" {
		minVer, err := semver.NewVersion(migration.MinRuntimeVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid min_runtime_version in migration %s: %w", version, err)
		}

		if runtimeVer.LessThan(minVer) {
			status.IsCompatible = false
			status.Status = "incompatible"
			status.Message = fmt.Sprintf("Runtime version %s is below minimum required %s", runtimeVersion, migration.MinRuntimeVersion)
			return status, nil
		}
	}

	// Check maximum runtime version
	if migration.MaxRuntimeVersion != "" {
		maxVer, err := semver.NewVersion(migration.MaxRuntimeVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid max_runtime_version in migration %s: %w", version, err)
		}

		if runtimeVer.GreaterThan(maxVer) {
			status.IsCompatible = false
			status.Status = "incompatible"
			status.Message = fmt.Sprintf("Runtime version %s exceeds maximum allowed %s", runtimeVersion, migration.MaxRuntimeVersion)
			return status, nil
		}
	}

	// Check if runtime version is significantly newer (warning)
	if migration.MinRuntimeVersion != "" {
		minVer, _ := semver.NewVersion(migration.MinRuntimeVersion)
		if runtimeVer.Major() > minVer.Major() {
			status.IsCompatible = true
			status.Status = "warning"
			status.Message = fmt.Sprintf("Runtime version %s is significantly newer than minimum required %s", runtimeVersion, migration.MinRuntimeVersion)
			return status, nil
		}
	}

	status.IsCompatible = true
	status.Status = "compatible"
	status.Message = "Migration is compatible with current runtime version"
	return status, nil
}

// GetMigrationPath returns the file paths for a migration
func GetMigrationPath(baseDir, version string) (upPath, downPath string) {
	// Find the migration directory by version
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this directory contains the version
		if strings.Contains(entry.Name(), version) {
			migrationDir := filepath.Join(baseDir, entry.Name())
			upPath = filepath.Join(migrationDir, "up.sql")
			downPath = filepath.Join(migrationDir, "down.sql")
			return upPath, downPath
		}
	}

	return "", ""
}
