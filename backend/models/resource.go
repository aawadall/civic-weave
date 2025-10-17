package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Resource represents a file or link in the resource library
type Resource struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Title         string     `json:"title" db:"title"`
	Description   string     `json:"description" db:"description"`
	ResourceType  string     `json:"resource_type" db:"resource_type"`
	FileURL       string     `json:"file_url" db:"file_url"`
	FileSize      *int64     `json:"file_size,omitempty" db:"file_size"`
	MimeType      *string    `json:"mime_type,omitempty" db:"mime_type"`
	Scope         string     `json:"scope" db:"scope"`
	ProjectID     *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	UploadedByID  uuid.UUID  `json:"uploaded_by_id" db:"uploaded_by_id"`
	Tags          []string   `json:"tags" db:"tags"`
	DownloadCount int        `json:"download_count" db:"download_count"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ResourceWithUploader includes resource and uploader info
type ResourceWithUploader struct {
	Resource
	UploaderName  string  `json:"uploader_name"`
	UploaderEmail string  `json:"uploader_email"`
	ProjectTitle  *string `json:"project_title,omitempty"`
}

// ResourceStats represents statistics for resources
type ResourceStats struct {
	TotalResources   int `json:"total_resources"`
	GlobalResources  int `json:"global_resources"`
	ProjectResources int `json:"project_resources"`
	TotalDownloads   int `json:"total_downloads"`
	RecentUploads    int `json:"recent_uploads"`
}

// ResourceFilters represents filters for resource queries
type ResourceFilters struct {
	Scope     *string    `json:"scope,omitempty"`
	ProjectID *uuid.UUID `json:"project_id,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Search    *string    `json:"search,omitempty"`
	Type      *string    `json:"type,omitempty"`
}

// ResourceService handles resource operations
type ResourceService struct {
	db *sql.DB
}

// NewResourceService creates a new resource service
func NewResourceService(db *sql.DB) *ResourceService {
	return &ResourceService{db: db}
}

// Create creates a new resource
func (s *ResourceService) Create(resource *Resource) error {
	resource.ID = uuid.New()
	tagsJSON, err := ToJSONArray(resource.Tags)
	if err != nil {
		return err
	}

	return s.db.QueryRow(resourceCreateQuery, resource.ID, resource.Title, resource.Description,
		resource.ResourceType, resource.FileURL, resource.FileSize, resource.MimeType,
		resource.Scope, resource.ProjectID, resource.UploadedByID, tagsJSON).
		Scan(&resource.CreatedAt, &resource.UpdatedAt)
}

// GetByID retrieves a resource by ID
func (s *ResourceService) GetByID(id uuid.UUID) (*Resource, error) {
	resource := &Resource{}
	var tagsJSON string

	err := s.db.QueryRow(resourceGetByIDQuery, id).Scan(
		&resource.ID, &resource.Title, &resource.Description, &resource.ResourceType,
		&resource.FileURL, &resource.FileSize, &resource.MimeType, &resource.Scope,
		&resource.ProjectID, &resource.UploadedByID, &tagsJSON, &resource.DownloadCount,
		&resource.CreatedAt, &resource.UpdatedAt, &resource.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse tags JSON
	if err := ParseJSONArray(tagsJSON, &resource.Tags); err != nil {
		return nil, err
	}

	return resource, nil
}

// List retrieves resources with filters
func (s *ResourceService) List(filters ResourceFilters, limit, offset int) ([]ResourceWithUploader, error) {
	// Build dynamic query based on filters
	query := resourceListQuery
	args := []interface{}{}
	argIndex := 1

	// Add WHERE conditions based on filters
	whereConditions := []string{"r.deleted_at IS NULL"}

	if filters.Scope != nil {
		whereConditions = append(whereConditions, "r.scope = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *filters.Scope)
		argIndex++
	}

	if filters.ProjectID != nil {
		whereConditions = append(whereConditions, "r.project_id = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *filters.ProjectID)
		argIndex++
	}

	if filters.Type != nil {
		whereConditions = append(whereConditions, "r.resource_type = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *filters.Type)
		argIndex++
	}

	if filters.Search != nil {
		whereConditions = append(whereConditions, "(r.title ILIKE $"+fmt.Sprintf("%d", argIndex)+" OR r.description ILIKE $"+fmt.Sprintf("%d", argIndex)+")")
		searchTerm := "%" + *filters.Search + "%"
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	if len(filters.Tags) > 0 {
		whereConditions = append(whereConditions, "r.tags ?| $"+fmt.Sprintf("%d", argIndex))
		args = append(args, filters.Tags)
		argIndex++
	}

	// Add WHERE clause to query
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add ORDER BY and LIMIT
	query += " ORDER BY r.created_at DESC LIMIT $" + fmt.Sprintf("%d", argIndex) + " OFFSET $" + fmt.Sprintf("%d", argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []ResourceWithUploader
	for rows.Next() {
		var resource ResourceWithUploader
		var tagsJSON string

		err := rows.Scan(
			&resource.ID, &resource.Title, &resource.Description, &resource.ResourceType,
			&resource.FileURL, &resource.FileSize, &resource.MimeType, &resource.Scope,
			&resource.ProjectID, &resource.UploadedByID, &tagsJSON, &resource.DownloadCount,
			&resource.CreatedAt, &resource.UpdatedAt, &resource.DeletedAt,
			&resource.UploaderName, &resource.UploaderEmail, &resource.ProjectTitle,
		)
		if err != nil {
			return nil, err
		}

		// Parse tags JSON
		if err := ParseJSONArray(tagsJSON, &resource.Tags); err != nil {
			return nil, err
		}

		resources = append(resources, resource)
	}

	return resources, rows.Err()
}

// Update updates a resource
func (s *ResourceService) Update(resource *Resource) error {
	tagsJSON, err := ToJSONArray(resource.Tags)
	if err != nil {
		return err
	}

	return s.db.QueryRow(resourceUpdateQuery, resource.ID, resource.Title, resource.Description,
		resource.ResourceType, resource.FileURL, resource.FileSize, resource.MimeType,
		resource.Scope, resource.ProjectID, tagsJSON).
		Scan(&resource.UpdatedAt)
}

// SoftDelete soft-deletes a resource
func (s *ResourceService) SoftDelete(id uuid.UUID) error {
	result, err := s.db.Exec(resourceSoftDeleteQuery, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// IncrementDownloadCount increments the download count for a resource
func (s *ResourceService) IncrementDownloadCount(id uuid.UUID) error {
	_, err := s.db.Exec(resourceIncrementDownloadQuery, id)
	return err
}

// GetStats returns resource statistics
func (s *ResourceService) GetStats() (*ResourceStats, error) {
	stats := &ResourceStats{}
	err := s.db.QueryRow(resourceGetStatsQuery).Scan(
		&stats.TotalResources, &stats.GlobalResources, &stats.ProjectResources,
		&stats.TotalDownloads, &stats.RecentUploads,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetRecentResources returns recently uploaded resources
func (s *ResourceService) GetRecentResources(limit int) ([]ResourceWithUploader, error) {
	rows, err := s.db.Query(resourceGetRecentQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []ResourceWithUploader
	for rows.Next() {
		var resource ResourceWithUploader
		var tagsJSON string

		err := rows.Scan(
			&resource.ID, &resource.Title, &resource.Description, &resource.ResourceType,
			&resource.FileURL, &resource.FileSize, &resource.MimeType, &resource.Scope,
			&resource.ProjectID, &resource.UploadedByID, &tagsJSON, &resource.DownloadCount,
			&resource.CreatedAt, &resource.UpdatedAt, &resource.DeletedAt,
			&resource.UploaderName, &resource.UploaderEmail, &resource.ProjectTitle,
		)
		if err != nil {
			return nil, err
		}

		// Parse tags JSON
		if err := ParseJSONArray(tagsJSON, &resource.Tags); err != nil {
			return nil, err
		}

		resources = append(resources, resource)
	}

	return resources, rows.Err()
}

// IsUploader checks if a user is the uploader of a resource
func (s *ResourceService) IsUploader(resourceID, userID uuid.UUID) (bool, error) {
	var count int
	err := s.db.QueryRow(resourceIsUploaderQuery, resourceID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetDB returns the database connection (needed for creating other services)
func (s *ResourceService) GetDB() *sql.DB {
	return s.db
}
