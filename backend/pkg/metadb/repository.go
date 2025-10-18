package metadb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"civicweave/backend/proto/dbagent"

	_ "github.com/lib/pq"
)

// Repository handles all metadata database operations
type Repository struct {
	db *sql.DB
}

// Database represents a managed database
type Database struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	ConnectionStringEnc string    `json:"connection_string_encrypted"`
	Description         string    `json:"description"`
	Environment         string    `json:"environment"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	CreatedBy           string    `json:"created_by"`
	Tags                []string  `json:"tags"`
}

// Deployment represents a deployment version
type Deployment struct {
	ID              string     `json:"id"`
	DatabaseID      string     `json:"database_id"`
	Version         string     `json:"version"`
	ManifestVersion string     `json:"manifest_version"`
	Status          string     `json:"status"`
	AppliedAt       *time.Time `json:"applied_at"`
	AppliedBy       string     `json:"applied_by"`
	Checksum        string     `json:"checksum"`
	ExecutionTimeMs int        `json:"execution_time_ms"`
	ErrorMessage    string     `json:"error_message"`
	DryRun          bool       `json:"dry_run"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Migration represents an individual migration execution
type Migration struct {
	ID              string     `json:"id"`
	DeploymentID    string     `json:"deployment_id"`
	Version         string     `json:"version"`
	Name            string     `json:"name"`
	Checksum        string     `json:"checksum"`
	Status          string     `json:"status"`
	ExecutionTimeMs int        `json:"execution_time_ms"`
	ErrorMessage    string     `json:"error_message"`
	AppliedAt       *time.Time `json:"applied_at"`
	RollbackAt      *time.Time `json:"rollback_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID                string          `json:"id"`
	DatabaseID        *string         `json:"database_id"`
	DeploymentID      *string         `json:"deployment_id"`
	Action            string          `json:"action"`
	UserAgent         string          `json:"user_agent"`
	ClientIP          string          `json:"client_ip"`
	RequestID         string          `json:"request_id"`
	StatusCode        int             `json:"status_code"`
	ErrorMessage      string          `json:"error_message"`
	ExecutionTimeMs   int             `json:"execution_time_ms"`
	RequestSizeBytes  int             `json:"request_size_bytes"`
	ResponseSizeBytes int             `json:"response_size_bytes"`
	CreatedAt         time.Time       `json:"created_at"`
	Metadata          json.RawMessage `json:"metadata"`
}

// NewRepository creates a new metadata database repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// InitializeSchema initializes the metadata database schema
func (r *Repository) InitializeSchema() error {
	// Read and execute schema SQL
	schemaSQL := `
		-- Check if tables exist, if not create them
		CREATE TABLE IF NOT EXISTS databases (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL UNIQUE,
			connection_string_encrypted TEXT NOT NULL,
			description TEXT,
			environment VARCHAR(50) DEFAULT 'production',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(255),
			tags TEXT[] DEFAULT '{}'
		);

		CREATE TABLE IF NOT EXISTS deployments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			database_id UUID NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
			version VARCHAR(50) NOT NULL,
			manifest_version VARCHAR(50),
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			applied_at TIMESTAMP,
			applied_by VARCHAR(255),
			checksum VARCHAR(64),
			execution_time_ms INTEGER,
			error_message TEXT,
			dry_run BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(database_id, version)
		);

		CREATE TABLE IF NOT EXISTS migrations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
			version VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			checksum VARCHAR(64) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			execution_time_ms INTEGER,
			error_message TEXT,
			applied_at TIMESTAMP,
			rollback_at TIMESTAMP,
			UNIQUE(deployment_id, version)
		);

		CREATE TABLE IF NOT EXISTS audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			database_id UUID REFERENCES databases(id) ON DELETE SET NULL,
			deployment_id UUID REFERENCES deployments(id) ON DELETE SET NULL,
			action VARCHAR(50) NOT NULL,
			user_agent VARCHAR(255),
			client_ip INET,
			request_id UUID,
			status_code INTEGER,
			error_message TEXT,
			execution_time_ms INTEGER,
			request_size_bytes INTEGER,
			response_size_bytes INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			metadata JSONB DEFAULT '{}'
		);

		CREATE TABLE IF NOT EXISTS api_keys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			key_id VARCHAR(255) NOT NULL UNIQUE,
			public_key_hash VARCHAR(64) NOT NULL,
			description TEXT,
			permissions TEXT[] DEFAULT '{}',
			is_active BOOLEAN DEFAULT true,
			expires_at TIMESTAMP,
			last_used_at TIMESTAMP,
			usage_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(255)
		);
	`

	if _, err := r.db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// RegisterDatabase registers a new database for management
func (r *Repository) RegisterDatabase(name, connectionString, description, environment, createdBy string, tags []string) (*Database, error) {
	query := `
		INSERT INTO databases (name, connection_string_encrypted, description, environment, created_by, tags)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	var id string
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(query, name, connectionString, description, environment, createdBy, tags).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to register database: %w", err)
	}

	return &Database{
		ID:                  id,
		Name:                name,
		ConnectionStringEnc: connectionString,
		Description:         description,
		Environment:         environment,
		IsActive:            true,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		CreatedBy:           createdBy,
		Tags:                tags,
	}, nil
}

// GetDatabase retrieves a database by name
func (r *Repository) GetDatabase(name string) (*Database, error) {
	query := `
		SELECT id, name, connection_string_encrypted, description, environment, 
		       is_active, created_at, updated_at, created_by, tags
		FROM databases 
		WHERE name = $1 AND is_active = true
	`

	var db Database
	var tags string
	err := r.db.QueryRow(query, name).Scan(
		&db.ID, &db.Name, &db.ConnectionStringEnc, &db.Description,
		&db.Environment, &db.IsActive, &db.CreatedAt, &db.UpdatedAt,
		&db.CreatedBy, &tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("database not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// Parse tags
	if err := json.Unmarshal([]byte(tags), &db.Tags); err != nil {
		db.Tags = []string{}
	}

	return &db, nil
}

// CreateDeployment creates a new deployment record
func (r *Repository) CreateDeployment(databaseID, version, manifestVersion, appliedBy, checksum string, dryRun bool) (*Deployment, error) {
	query := `
		INSERT INTO deployments (database_id, version, manifest_version, applied_by, checksum, dry_run)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	var id string
	var createdAt time.Time
	err := r.db.QueryRow(query, databaseID, version, manifestVersion, appliedBy, checksum, dryRun).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	return &Deployment{
		ID:              id,
		DatabaseID:      databaseID,
		Version:         version,
		ManifestVersion: manifestVersion,
		Status:          "pending",
		AppliedBy:       appliedBy,
		Checksum:        checksum,
		DryRun:          dryRun,
		CreatedAt:       createdAt,
	}, nil
}

// UpdateDeploymentStatus updates the status of a deployment
func (r *Repository) UpdateDeploymentStatus(deploymentID, status string, executionTimeMs int, errorMessage string) error {
	query := `
		UPDATE deployments 
		SET status = $2, execution_time_ms = $3, error_message = $4, applied_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.db.Exec(query, deploymentID, status, executionTimeMs, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

// AddMigration adds a migration record to a deployment
func (r *Repository) AddMigration(deploymentID, version, name, checksum string) (*Migration, error) {
	query := `
		INSERT INTO migrations (deployment_id, version, name, checksum)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	var id string
	var createdAt time.Time
	err := r.db.QueryRow(query, deploymentID, version, name, checksum).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add migration: %w", err)
	}

	return &Migration{
		ID:           id,
		DeploymentID: deploymentID,
		Version:      version,
		Name:         name,
		Checksum:     checksum,
		Status:       "pending",
	}, nil
}

// UpdateMigrationStatus updates the status of a migration
func (r *Repository) UpdateMigrationStatus(migrationID, status string, executionTimeMs int, errorMessage string) error {
	query := `
		UPDATE migrations 
		SET status = $2, execution_time_ms = $3, error_message = $4, applied_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.db.Exec(query, migrationID, status, executionTimeMs, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to update migration status: %w", err)
	}

	return nil
}

// GetDeploymentHistory retrieves deployment history for a database
func (r *Repository) GetDeploymentHistory(databaseName string, limit, offset int) ([]*dbagent.DeploymentVersion, error) {
	query := `
		SELECT 
			d.id, d.version, d.status, d.applied_at, d.applied_by, 
			d.execution_time_ms, d.checksum
		FROM deployments d
		JOIN databases db ON d.database_id = db.id
		WHERE db.name = $1
		ORDER BY d.applied_at DESC NULLS LAST, d.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, databaseName, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployment history: %w", err)
	}
	defer rows.Close()

	var deployments []*dbagent.DeploymentVersion
	for rows.Next() {
		var dep dbagent.DeploymentVersion
		var appliedAt *time.Time

		err := rows.Scan(
			&dep.Id, &dep.Version, &dep.Status, &appliedAt, &dep.AppliedBy,
			&dep.ExecutionTimeMs, &dep.Checksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}

		if appliedAt != nil {
			dep.AppliedAt = appliedAt.Unix()
		}

		deployments = append(deployments, &dep)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployment rows: %w", err)
	}

	return deployments, nil
}

// GetCurrentDatabaseVersion gets the current version of a database
func (r *Repository) GetCurrentDatabaseVersion(databaseName string) (string, error) {
	query := `
		SELECT d.version
		FROM deployments d
		JOIN databases db ON d.database_id = db.id
		WHERE db.name = $1 AND d.status = 'applied'
		ORDER BY d.applied_at DESC
		LIMIT 1
	`

	var version string
	err := r.db.QueryRow(query, databaseName).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return "0.0.0", nil
		}
		return "", fmt.Errorf("failed to get current database version: %w", err)
	}

	return version, nil
}

// LogAuditEntry logs an audit entry
func (r *Repository) LogAuditEntry(entry *AuditLog) error {
	query := `
		INSERT INTO audit_log (
			database_id, deployment_id, action, user_agent, client_ip, request_id,
			status_code, error_message, execution_time_ms, request_size_bytes,
			response_size_bytes, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	var databaseID, deploymentID *string
	if entry.DatabaseID != nil {
		databaseID = entry.DatabaseID
	}
	if entry.DeploymentID != nil {
		deploymentID = entry.DeploymentID
	}

	_, err := r.db.Exec(query,
		databaseID, deploymentID, entry.Action, entry.UserAgent, entry.ClientIP,
		entry.RequestID, entry.StatusCode, entry.ErrorMessage, entry.ExecutionTimeMs,
		entry.RequestSizeBytes, entry.ResponseSizeBytes, entry.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to log audit entry: %w", err)
	}

	return nil
}

// ValidateAPIKey validates an API key
func (r *Repository) ValidateAPIKey(keyID, publicKey string) (bool, error) {
	query := `
		SELECT is_active, expires_at
		FROM api_keys
		WHERE key_id = $1 AND public_key_hash = $2
	`

	var isActive bool
	var expiresAt *time.Time
	err := r.db.QueryRow(query, keyID, publicKey).Scan(&isActive, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to validate API key: %w", err)
	}

	if !isActive {
		return false, nil
	}

	if expiresAt != nil && time.Now().After(*expiresAt) {
		return false, nil
	}

	// Update last used timestamp
	updateQuery := `
		UPDATE api_keys 
		SET last_used_at = CURRENT_TIMESTAMP, usage_count = usage_count + 1
		WHERE key_id = $1
	`
	r.db.Exec(updateQuery, keyID)

	return true, nil
}

// RegisterAPIKey registers a new API key
func (r *Repository) RegisterAPIKey(keyID, publicKeyHash, description, createdBy string, permissions []string, expiresAt *time.Time) error {
	query := `
		INSERT INTO api_keys (key_id, public_key_hash, description, created_by, permissions, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(query, keyID, publicKeyHash, description, createdBy, permissions, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to register API key: %w", err)
	}

	return nil
}

// GetDeploymentByID retrieves a deployment by ID
func (r *Repository) GetDeploymentByID(deploymentID string) (*Deployment, error) {
	query := `
		SELECT id, database_id, version, manifest_version, status, applied_at, 
		       applied_by, checksum, execution_time_ms, error_message, dry_run, created_at
		FROM deployments 
		WHERE id = $1
	`

	var dep Deployment
	var appliedAt *time.Time
	err := r.db.QueryRow(query, deploymentID).Scan(
		&dep.ID, &dep.DatabaseID, &dep.Version, &dep.ManifestVersion, &dep.Status,
		&appliedAt, &dep.AppliedBy, &dep.Checksum, &dep.ExecutionTimeMs,
		&dep.ErrorMessage, &dep.DryRun, &dep.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deployment not found: %s", deploymentID)
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	if appliedAt != nil {
		dep.AppliedAt = appliedAt
	}

	return &dep, nil
}

// GetMigrationsForDeployment retrieves all migrations for a deployment
func (r *Repository) GetMigrationsForDeployment(deploymentID string) ([]*Migration, error) {
	query := `
		SELECT id, deployment_id, version, name, checksum, status, execution_time_ms,
		       error_message, applied_at, rollback_at
		FROM migrations
		WHERE deployment_id = $1
		ORDER BY version
	`

	rows, err := r.db.Query(query, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var migrations []*Migration
	for rows.Next() {
		var migration Migration
		var appliedAt, rollbackAt *time.Time

		err := rows.Scan(
			&migration.ID, &migration.DeploymentID, &migration.Version, &migration.Name,
			&migration.Checksum, &migration.Status, &migration.ExecutionTimeMs,
			&migration.ErrorMessage, &appliedAt, &rollbackAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}

		if appliedAt != nil {
			migration.AppliedAt = appliedAt
		}
		if rollbackAt != nil {
			migration.RollbackAt = rollbackAt
		}

		migrations = append(migrations, &migration)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration rows: %w", err)
	}

	return migrations, nil
}

// Close closes the database connection
func (r *Repository) Close() error {
	return r.db.Close()
}
