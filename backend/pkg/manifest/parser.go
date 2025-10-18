package manifest

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"civicweave/backend/proto/dbagent"
)

// Parser handles parsing of manifest files and directories
type Parser struct {
	manifestPath string
}

// NewParser creates a new manifest parser
func NewParser(manifestPath string) *Parser {
	return &Parser{
		manifestPath: manifestPath,
	}
}

// ParseManifest parses a manifest directory and returns a Manifest protobuf message
func (p *Parser) ParseManifest() (*dbagent.Manifest, error) {
	// Check if manifest path exists
	if _, err := os.Stat(p.manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest directory does not exist: %s", p.manifestPath)
	}

	manifest := &dbagent.Manifest{
		Version:     "1.0.0",
		Description: "Database Manifest",
		Author:      "system",
		CreatedAt:   time.Now().Unix(),
		Migrations:  []*dbagent.Migration{},
		SeedData:    []*dbagent.SeedData{},
		Metadata: &dbagent.ManifestMetadata{
			MinRuntimeVersion: "1.0.0",
			Tags:              []string{},
			CustomProperties:  make(map[string]string),
		},
	}

	// Parse migrations
	migrations, err := p.parseMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to parse migrations: %w", err)
	}
	manifest.Migrations = migrations

	// Parse seed data
	seedData, err := p.parseSeedData()
	if err != nil {
		return nil, fmt.Errorf("failed to parse seed data: %w", err)
	}
	manifest.SeedData = seedData

	// Parse metadata.json if it exists
	metadataPath := filepath.Join(p.manifestPath, "metadata.json")
	if _, err := os.Stat(metadataPath); err == nil {
		metadata, err := p.parseMetadata(metadataPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
		if metadata != nil {
			manifest.Metadata = metadata
		}
	}

	return manifest, nil
}

// parseMigrations parses migration files from the migrations directory
func (p *Parser) parseMigrations() ([]*dbagent.Migration, error) {
	migrationsDir := filepath.Join(p.manifestPath, "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return []*dbagent.Migration{}, nil
	}

	var migrations []*dbagent.Migration
	versionRegex := regexp.MustCompile(`^V(\d+)__(.+)$`)

	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".sql") {
			return nil
		}

		// Parse filename to extract version and name
		matches := versionRegex.FindStringSubmatch(d.Name())
		if len(matches) != 3 {
			return fmt.Errorf("invalid migration filename format: %s (expected V###__description.sql)", d.Name())
		}

		version := matches[1]
		name := strings.ReplaceAll(matches[2], "_", " ")

		// Read migration file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", d.Name(), err)
		}

		// Parse UP and DOWN sections
		upSQL, downSQL, err := p.parseMigrationSections(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse migration sections in %s: %w", d.Name(), err)
		}

		// Calculate checksum
		checksum := p.calculateChecksum(content)

		migration := &dbagent.Migration{
			Version:         fmt.Sprintf("V%s", version),
			Name:            name,
			Description:     fmt.Sprintf("Migration %s: %s", version, name),
			UpSql:           upSQL,
			DownSql:         downSQL,
			Dependencies:    []string{},
			Checksum:        checksum,
			ExecutionTimeMs: 0,
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationSections parses UP and DOWN sections from migration content
func (p *Parser) parseMigrationSections(content string) (string, string, error) {
	// Split by -- DOWN marker
	sections := strings.Split(content, "-- DOWN")
	if len(sections) < 2 {
		return content, "", fmt.Errorf("migration must contain -- UP and -- DOWN sections")
	}

	upSection := strings.TrimSpace(sections[0])
	downSection := strings.TrimSpace(sections[1])

	// Remove -- UP header if present
	if strings.HasPrefix(upSection, "-- UP") {
		upSection = strings.TrimSpace(strings.TrimPrefix(upSection, "-- UP"))
	}

	// Remove -- DOWN header if present
	if strings.HasPrefix(downSection, "-- DOWN") {
		downSection = strings.TrimSpace(strings.TrimPrefix(downSection, "-- DOWN"))
	}

	return upSection, downSection, nil
}

// parseSeedData parses seed data files from the seeds directory
func (p *Parser) parseSeedData() ([]*dbagent.SeedData, error) {
	seedsDir := filepath.Join(p.manifestPath, "seeds")
	if _, err := os.Stat(seedsDir); os.IsNotExist(err) {
		return []*dbagent.SeedData{}, nil
	}

	var seedDataList []*dbagent.SeedData

	err := filepath.WalkDir(seedsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Skip non-SQL files
		if !strings.HasSuffix(d.Name(), ".sql") {
			return nil
		}

		// Extract environment from directory structure
		relPath, err := filepath.Rel(seedsDir, path)
		if err != nil {
			return err
		}

		pathParts := strings.Split(relPath, string(filepath.Separator))
		environment := "default"
		if len(pathParts) > 1 {
			environment = pathParts[0] // First directory is environment
		}

		// Extract table name from filename
		tableName := strings.TrimSuffix(filepath.Base(path), ".sql")

		// Read seed file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read seed file %s: %w", path, err)
		}

		// Split content into SQL statements
		statements := p.parseSQLStatements(string(content))

		// Calculate checksum
		checksum := p.calculateChecksum(content)

		seedData := &dbagent.SeedData{
			Environment:   environment,
			TableName:     tableName,
			SqlStatements: statements,
			Checksum:      checksum,
		}

		seedDataList = append(seedDataList, seedData)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return seedDataList, nil
}

// parseSQLStatements splits SQL content into individual statements
func (p *Parser) parseSQLStatements(content string) []string {
	// Remove comments and normalize whitespace
	content = p.normalizeSQL(content)

	// Split by semicolon
	statements := strings.Split(content, ";")
	var result []string

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}

// normalizeSQL normalizes SQL content by removing comments and extra whitespace
func (p *Parser) normalizeSQL(content string) string {
	// Remove single-line comments
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "--") && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}

	return strings.Join(result, " ")
}

// parseMetadata parses the metadata.json file
func (p *Parser) parseMetadata(metadataPath string) (*dbagent.ManifestMetadata, error) {
	content, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var metadata struct {
		Version           string            `json:"version"`
		Description       string            `json:"description"`
		Author            string            `json:"author"`
		MinRuntimeVersion string            `json:"min_runtime_version"`
		MaxRuntimeVersion string            `json:"max_runtime_version"`
		Tags              []string          `json:"tags"`
		CustomProperties  map[string]string `json:"custom_properties"`
	}

	if err := json.Unmarshal(content, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	return &dbagent.ManifestMetadata{
		MinRuntimeVersion: metadata.MinRuntimeVersion,
		MaxRuntimeVersion: metadata.MaxRuntimeVersion,
		Tags:              metadata.Tags,
		CustomProperties:  metadata.CustomProperties,
	}, nil
}

// calculateChecksum calculates SHA256 checksum of content
func (p *Parser) calculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// ValidateManifest validates a manifest for consistency and completeness
func (p *Parser) ValidateManifest(manifest *dbagent.Manifest) error {
	// Validate version format
	if manifest.Version == "" {
		return fmt.Errorf("manifest version is required")
	}

	// Validate migrations
	for _, migration := range manifest.Migrations {
		if err := p.validateMigration(migration); err != nil {
			return fmt.Errorf("invalid migration %s: %w", migration.Version, err)
		}
	}

	// Validate seed data
	for _, seedData := range manifest.SeedData {
		if err := p.validateSeedData(seedData); err != nil {
			return fmt.Errorf("invalid seed data for table %s: %w", seedData.TableName, err)
		}
	}

	// Check for duplicate migration versions
	versions := make(map[string]bool)
	for _, migration := range manifest.Migrations {
		if versions[migration.Version] {
			return fmt.Errorf("duplicate migration version: %s", migration.Version)
		}
		versions[migration.Version] = true
	}

	return nil
}

// validateMigration validates a single migration
func (p *Parser) validateMigration(migration *dbagent.Migration) error {
	if migration.Version == "" {
		return fmt.Errorf("migration version is required")
	}

	if migration.Name == "" {
		return fmt.Errorf("migration name is required")
	}

	if migration.UpSql == "" {
		return fmt.Errorf("migration UP SQL is required")
	}

	if migration.Checksum == "" {
		return fmt.Errorf("migration checksum is required")
	}

	return nil
}

// validateSeedData validates seed data
func (p *Parser) validateSeedData(seedData *dbagent.SeedData) error {
	if seedData.TableName == "" {
		return fmt.Errorf("table name is required")
	}

	if seedData.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	if len(seedData.SqlStatements) == 0 {
		return fmt.Errorf("at least one SQL statement is required")
	}

	return nil
}

// WriteManifest writes a manifest to a directory
func (p *Parser) WriteManifest(manifest *dbagent.Manifest) error {
	// Create manifest directory
	if err := os.MkdirAll(p.manifestPath, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// Write migrations
	if err := p.writeMigrations(manifest.Migrations); err != nil {
		return fmt.Errorf("failed to write migrations: %w", err)
	}

	// Write seed data
	if err := p.writeSeedData(manifest.SeedData); err != nil {
		return fmt.Errorf("failed to write seed data: %w", err)
	}

	// Write metadata
	if err := p.writeMetadata(manifest.Metadata); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// writeMigrations writes migration files to the migrations directory
func (p *Parser) writeMigrations(migrations []*dbagent.Migration) error {
	migrationsDir := filepath.Join(p.manifestPath, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return err
	}

	for _, migration := range migrations {
		// Create migration filename
		version := strings.TrimPrefix(migration.Version, "V")
		filename := fmt.Sprintf("V%s__%s.sql", version, strings.ReplaceAll(migration.Name, " ", "_"))
		filepath := filepath.Join(migrationsDir, filename)

		// Create migration content
		content := fmt.Sprintf("-- UP\n%s\n\n-- DOWN\n%s", migration.UpSql, migration.DownSql)

		// Write migration file
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write migration file %s: %w", filename, err)
		}
	}

	return nil
}

// writeSeedData writes seed data files to the seeds directory
func (p *Parser) writeSeedData(seedDataList []*dbagent.SeedData) error {
	seedsDir := filepath.Join(p.manifestPath, "seeds")
	if err := os.MkdirAll(seedsDir, 0755); err != nil {
		return err
	}

	for _, seedData := range seedDataList {
		// Create environment directory
		envDir := filepath.Join(seedsDir, seedData.Environment)
		if err := os.MkdirAll(envDir, 0755); err != nil {
			return err
		}

		// Create seed file
		filename := fmt.Sprintf("%s.sql", seedData.TableName)
		filepath := filepath.Join(envDir, filename)

		// Create seed content
		content := strings.Join(seedData.SqlStatements, ";\n") + ";"

		// Write seed file
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write seed file %s: %w", filename, err)
		}
	}

	return nil
}

// writeMetadata writes metadata.json file
func (p *Parser) writeMetadata(metadata *dbagent.ManifestMetadata) error {
	if metadata == nil {
		return nil
	}

	metadataFile := filepath.Join(p.manifestPath, "metadata.json")

	metadataData := map[string]interface{}{
		"min_runtime_version": metadata.MinRuntimeVersion,
		"max_runtime_version": metadata.MaxRuntimeVersion,
		"tags":                metadata.Tags,
		"custom_properties":   metadata.CustomProperties,
	}

	content, err := json.MarshalIndent(metadataData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}
