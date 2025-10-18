package dbagent

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"civicweave/backend/pkg/metadb"
	"civicweave/backend/proto/dbagent"

	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AgentService implements the gRPC DatabaseAgent service
type AgentService struct {
	metaRepo    *metadb.Repository
	auditLogger *AuditLogger
	dbagent.UnimplementedDatabaseAgentServer
}

// NewAgentService creates a new agent service
func NewAgentService(metaRepo *metadb.Repository, auditLogger *AuditLogger) *AgentService {
	return &AgentService{
		metaRepo:    metaRepo,
		auditLogger: auditLogger,
	}
}

// Ping implements the Ping gRPC method
func (s *AgentService) Ping(ctx context.Context, req *dbagent.PingRequest) (*dbagent.PingResponse, error) {
	startTime := time.Now()

	response := &dbagent.PingResponse{
		AgentVersion: "1.0.0",
		Status:       "healthy",
		Timestamp:    time.Now().Unix(),
	}

	// Log audit entry
	executionTime := int(time.Since(startTime).Milliseconds())
	s.auditLogger.LogRequest(ctx, "ping", nil, nil, 200, "", executionTime, 0, 0, map[string]interface{}{
		"client_version": req.ClientVersion,
	})

	return response, nil
}

// CompareManifest implements the CompareManifest gRPC method
func (s *AgentService) CompareManifest(ctx context.Context, req *dbagent.CompareManifestRequest) (*dbagent.CompareManifestResponse, error) {
	startTime := time.Now()

	// Get database information
	database, err := s.metaRepo.GetDatabase(req.DatabaseName)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "compare", &database.ID, nil, 404, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.NotFound, "database not found: %v", err)
	}

	// Connect to the target database
	targetDB, err := sql.Open("postgres", database.ConnectionStringEnc)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "compare", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "failed to connect to database: %v", err)
	}
	defer targetDB.Close()

	// Get current schema state
	currentSchema, err := s.getCurrentSchemaState(targetDB)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "compare", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "failed to get current schema: %v", err)
	}

	// Compare with manifest
	comparison, err := s.compareSchemaWithManifest(currentSchema, req.Manifest, req.IncludeDataDiff)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "compare", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "failed to compare schemas: %v", err)
	}

	// Log audit entry
	executionTime := int(time.Since(startTime).Milliseconds())
	s.auditLogger.LogRequest(ctx, "compare", &database.ID, nil, 200, "", executionTime, 0, 0, map[string]interface{}{
		"include_data_diff": req.IncludeDataDiff,
		"is_identical":      comparison.IsIdentical,
	})

	return comparison, nil
}

// DownloadManifest implements the DownloadManifest gRPC method
func (s *AgentService) DownloadManifest(ctx context.Context, req *dbagent.DownloadManifestRequest) (*dbagent.DownloadManifestResponse, error) {
	startTime := time.Now()

	// Get database information
	database, err := s.metaRepo.GetDatabase(req.DatabaseName)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "download", &database.ID, nil, 404, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.NotFound, "database not found: %v", err)
	}

	// Connect to the target database
	targetDB, err := sql.Open("postgres", database.ConnectionStringEnc)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "download", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "failed to connect to database: %v", err)
	}
	defer targetDB.Close()

	// Extract schema as manifest
	manifest, err := s.extractSchemaAsManifest(targetDB, req.IncludeData, req.Environment)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "download", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "failed to extract schema: %v", err)
	}

	// Calculate checksum
	checksum := s.calculateManifestChecksum(manifest)

	// Log audit entry
	executionTime := int(time.Since(startTime).Milliseconds())
	s.auditLogger.LogRequest(ctx, "download", &database.ID, nil, 200, "", executionTime, 0, 0, map[string]interface{}{
		"include_data":  req.IncludeData,
		"environment":   req.Environment,
		"objects_count": len(manifest.Migrations) + len(manifest.SeedData),
	})

	return &dbagent.DownloadManifestResponse{
		Manifest:     manifest,
		Checksum:     checksum,
		ObjectsCount: int32(len(manifest.Migrations) + len(manifest.SeedData)),
	}, nil
}

// DeployManifest implements the DeployManifest gRPC method
func (s *AgentService) DeployManifest(ctx context.Context, req *dbagent.DeployManifestRequest) (*dbagent.DeployManifestResponse, error) {
	startTime := time.Now()

	// Get database information
	database, err := s.metaRepo.GetDatabase(req.DatabaseName)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "deploy", &database.ID, nil, 404, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.NotFound, "database not found: %v", err)
	}

	// Create deployment record
	deployment, err := s.metaRepo.CreateDeployment(
		database.ID,
		req.Manifest.Version,
		req.Manifest.Version,
		"system", // TODO: Get from auth context
		s.calculateManifestChecksum(req.Manifest),
		req.DryRun,
	)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "deploy", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "failed to create deployment record: %v", err)
	}

	// Connect to the target database
	targetDB, err := sql.Open("postgres", database.ConnectionStringEnc)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "deploy", &database.ID, &deployment.ID, 500, err.Error(), executionTime, 0, 0, nil)
		s.metaRepo.UpdateDeploymentStatus(deployment.ID, "failed", 0, err.Error())
		return nil, status.Errorf(codes.Internal, "failed to connect to database: %v", err)
	}
	defer targetDB.Close()

	// Deploy the manifest
	result, err := s.deployManifestToDatabase(targetDB, req.Manifest, req.DryRun, req.TargetVersion, req.Force)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "deploy", &database.ID, &deployment.ID, 500, err.Error(), executionTime, 0, 0, nil)
		s.metaRepo.UpdateDeploymentStatus(deployment.ID, "failed", executionTime, err.Error())
		return nil, status.Errorf(codes.Internal, "deployment failed: %v", err)
	}

	// Update deployment status
	executionTime := int(time.Since(startTime).Milliseconds())
	status := "applied"
	if req.DryRun {
		status = "dry_run"
	}
	s.metaRepo.UpdateDeploymentStatus(deployment.ID, status, executionTime, "")

	// Log audit entry
	s.auditLogger.LogRequest(ctx, "deploy", &database.ID, &deployment.ID, 200, "", executionTime, 0, 0, map[string]interface{}{
		"dry_run":        req.DryRun,
		"target_version": req.TargetVersion,
		"force":          req.Force,
		"migrations":     len(result.Migrations),
	})

	return result, nil
}

// GetDeploymentHistory implements the GetDeploymentHistory gRPC method
func (s *AgentService) GetDeploymentHistory(ctx context.Context, req *dbagent.DeploymentHistoryRequest) (*dbagent.DeploymentHistoryResponse, error) {
	startTime := time.Now()

	// Get deployment history
	deployments, err := s.metaRepo.GetDeploymentHistory(req.DatabaseName, int(req.Limit), int(req.Offset))
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "history", nil, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "failed to get deployment history: %v", err)
	}

	// Check if there are more deployments
	hasMore := len(deployments) == int(req.Limit)

	// Log audit entry
	executionTime := int(time.Since(startTime).Milliseconds())
	s.auditLogger.LogRequest(ctx, "history", nil, nil, 200, "", executionTime, 0, 0, map[string]interface{}{
		"limit":  req.Limit,
		"offset": req.Offset,
		"count":  len(deployments),
	})

	return &dbagent.DeploymentHistoryResponse{
		Deployments: deployments,
		TotalCount:  int32(len(deployments)),
		HasMore:     hasMore,
	}, nil
}

// Bootstrap implements the Bootstrap gRPC method
func (s *AgentService) Bootstrap(ctx context.Context, req *dbagent.BootstrapRequest) (*dbagent.BootstrapResponse, error) {
	startTime := time.Now()

	// Register the database if it doesn't exist
	database, err := s.metaRepo.GetDatabase(req.DatabaseName)
	if err != nil {
		// Database doesn't exist, register it
		database, err = s.metaRepo.RegisterDatabase(
			req.DatabaseName,
			req.ConnectionString,
			"Bootstrap database",
			"production",
			"system", // TODO: Get from auth context
			[]string{"bootstrap"},
		)
		if err != nil {
			executionTime := int(time.Since(startTime).Milliseconds())
			s.auditLogger.LogRequest(ctx, "bootstrap", nil, nil, 500, err.Error(), executionTime, 0, 0, nil)
			return nil, status.Errorf(codes.Internal, "failed to register database: %v", err)
		}
	}

	// Create database if requested
	if req.CreateDatabase {
		err = s.createDatabase(req.ConnectionString, req.DatabaseName)
		if err != nil {
			executionTime := int(time.Since(startTime).Milliseconds())
			s.auditLogger.LogRequest(ctx, "bootstrap", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
			return nil, status.Errorf(codes.Internal, "failed to create database: %v", err)
		}
	}

	// Connect to the target database
	targetDB, err := sql.Open("postgres", req.ConnectionString)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "bootstrap", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "failed to connect to database: %v", err)
	}
	defer targetDB.Close()

	// Deploy the manifest
	result, err := s.deployManifestToDatabase(targetDB, req.Manifest, false, "", false)
	if err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		s.auditLogger.LogRequest(ctx, "bootstrap", &database.ID, nil, 500, err.Error(), executionTime, 0, 0, nil)
		return nil, status.Errorf(codes.Internal, "bootstrap failed: %v", err)
	}

	// Log audit entry
	executionTime := int(time.Since(startTime).Milliseconds())
	s.auditLogger.LogRequest(ctx, "bootstrap", &database.ID, nil, 200, "", executionTime, 0, 0, map[string]interface{}{
		"create_database": req.CreateDatabase,
		"migrations":      len(result.Migrations),
	})

	return &dbagent.BootstrapResponse{
		Success:         true,
		DatabaseName:    req.DatabaseName,
		Migrations:      result.Migrations,
		ExecutionTimeMs: int64(executionTime),
	}, nil
}

// Helper methods

// getCurrentSchemaState extracts the current schema state from a database
func (s *AgentService) getCurrentSchemaState(db *sql.DB) (map[string]interface{}, error) {
	// This is a simplified implementation
	// In a real implementation, you would query the database schema
	// and return a comprehensive schema representation

	schema := make(map[string]interface{})

	// Get tables
	tables := make([]string, 0)
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	schema["tables"] = tables

	return schema, nil
}

// compareSchemaWithManifest compares current schema with manifest
func (s *AgentService) compareSchemaWithManifest(currentSchema map[string]interface{}, manifest *dbagent.Manifest, includeDataDiff bool) (*dbagent.CompareManifestResponse, error) {
	// This is a simplified implementation
	// In a real implementation, you would perform detailed schema comparison

	response := &dbagent.CompareManifestResponse{
		IsIdentical:     true,
		Differences:     []string{},
		MissingObjects:  []string{},
		ExtraObjects:    []string{},
		DataDifferences: []*dbagent.DataDiff{},
		LocalChecksum:   s.calculateManifestChecksum(manifest),
		RemoteChecksum:  "remote_checksum_placeholder",
	}

	// For now, assume schemas are identical
	// In a real implementation, you would compare:
	// - Tables and their columns
	// - Indexes
	// - Constraints
	// - Functions and procedures
	// - Data differences (if requested)

	return response, nil
}

// extractSchemaAsManifest extracts current schema as a manifest
func (s *AgentService) extractSchemaAsManifest(db *sql.DB, includeData bool, environment string) (*dbagent.Manifest, error) {
	// This is a simplified implementation
	// In a real implementation, you would reverse-engineer the schema

	manifest := &dbagent.Manifest{
		Version:     "1.0.0",
		Description: "Extracted schema manifest",
		Author:      "system",
		CreatedAt:   time.Now().Unix(),
		Migrations:  []*dbagent.Migration{},
		SeedData:    []*dbagent.SeedData{},
		Metadata: &dbagent.ManifestMetadata{
			MinRuntimeVersion: "1.0.0",
			Tags:              []string{"extracted"},
			CustomProperties:  make(map[string]string),
		},
	}

	// Extract tables as migrations
	tables := make([]string, 0)
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	// Create a migration for each table (simplified)
	for i, tableName := range tables {
		migration := &dbagent.Migration{
			Version:         fmt.Sprintf("V%03d", i+1),
			Name:            fmt.Sprintf("create_%s_table", tableName),
			Description:     fmt.Sprintf("Create %s table", tableName),
			UpSql:           fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY);", tableName),
			DownSql:         fmt.Sprintf("DROP TABLE %s;", tableName),
			Dependencies:    []string{},
			Checksum:        fmt.Sprintf("checksum_%s", tableName),
			ExecutionTimeMs: 0,
		}
		manifest.Migrations = append(manifest.Migrations, migration)
	}

	return manifest, nil
}

// deployManifestToDatabase deploys a manifest to a database
func (s *AgentService) deployManifestToDatabase(db *sql.DB, manifest *dbagent.Manifest, dryRun bool, targetVersion string, force bool) (*dbagent.DeployManifestResponse, error) {
	response := &dbagent.DeployManifestResponse{
		Success:         true,
		Status:          "applied",
		Migrations:      []*dbagent.MigrationResult{},
		ExecutionPlan:   "Execution plan placeholder",
		ExecutionTimeMs: 0,
		Warnings:        []string{},
		Errors:          []string{},
	}

	if dryRun {
		response.Status = "dry_run"
	}

	// Apply each migration
	for _, migration := range manifest.Migrations {
		migrationResult := &dbagent.MigrationResult{
			Version:         migration.Version,
			Name:            migration.Name,
			Status:          "applied",
			ExecutionTimeMs: 0,
			Checksum:        migration.Checksum,
		}

		if !dryRun {
			// Execute migration
			startTime := time.Now()
			_, err := db.Exec(migration.UpSql)
			executionTime := int(time.Since(startTime).Milliseconds())

			migrationResult.ExecutionTimeMs = int64(executionTime)

			if err != nil {
				migrationResult.Status = "failed"
				migrationResult.ErrorMessage = err.Error()
				response.Errors = append(response.Errors, fmt.Sprintf("Migration %s failed: %v", migration.Version, err))
			}
		}

		response.Migrations = append(response.Migrations, migrationResult)
	}

	return response, nil
}

// calculateManifestChecksum calculates a checksum for a manifest
func (s *AgentService) calculateManifestChecksum(manifest *dbagent.Manifest) string {
	// This is a simplified checksum calculation
	// In a real implementation, you would hash the entire manifest content
	return fmt.Sprintf("manifest_%d", time.Now().Unix())
}

// createDatabase creates a new database
func (s *AgentService) createDatabase(connectionString, databaseName string) error {
	// This is a simplified implementation
	// In a real implementation, you would parse the connection string
	// and create the database using appropriate SQL commands

	log.Printf("Creating database: %s", databaseName)
	// For PostgreSQL, you would typically:
	// 1. Connect to the postgres database
	// 2. Execute CREATE DATABASE command
	// 3. Handle errors appropriately

	return nil
}
