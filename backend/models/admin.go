package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Admin represents an admin user in the system
type Admin struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AdminService handles admin operations
type AdminService struct {
	db *sql.DB
}

// NewAdminService creates a new admin service
func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{db: db}
}

// Create creates a new admin
func (s *AdminService) Create(admin *Admin) error {
	admin.ID = uuid.New()
	return s.db.QueryRow(adminCreateQuery, admin.ID, admin.UserID, admin.Name).
		Scan(&admin.CreatedAt)
}

// GetByID retrieves an admin by ID
func (s *AdminService) GetByID(id uuid.UUID) (*Admin, error) {
	admin := &Admin{}

	err := s.db.QueryRow(adminGetByIDQuery, id).Scan(
		&admin.ID, &admin.UserID, &admin.Name, &admin.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return admin, nil
}

// GetByUserID retrieves an admin by user ID
func (s *AdminService) GetByUserID(userID uuid.UUID) (*Admin, error) {
	admin := &Admin{}

	err := s.db.QueryRow(adminGetByUserIDQuery, userID).Scan(
		&admin.ID, &admin.UserID, &admin.Name, &admin.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return admin, nil
}

// List retrieves all admins
func (s *AdminService) List() ([]*Admin, error) {
	rows, err := s.db.Query(adminListQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []*Admin
	for rows.Next() {
		admin := &Admin{}
		err := rows.Scan(
			&admin.ID, &admin.UserID, &admin.Name, &admin.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}

	return admins, rows.Err()
}

// Update updates an admin
func (s *AdminService) Update(admin *Admin) error {
	_, err := s.db.Exec(adminUpdateQuery, admin.ID, admin.Name)
	return err
}

// Delete deletes an admin
func (s *AdminService) Delete(id uuid.UUID) error {
	_, err := s.db.Exec(adminDeleteQuery, id)
	return err
}
