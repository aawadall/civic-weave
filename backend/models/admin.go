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
	query := `
		INSERT INTO admins (id, user_id, name)
		VALUES ($1, $2, $3)
		RETURNING created_at`

	admin.ID = uuid.New()
	return s.db.QueryRow(query, admin.ID, admin.UserID, admin.Name).
		Scan(&admin.CreatedAt)
}

// GetByID retrieves an admin by ID
func (s *AdminService) GetByID(id uuid.UUID) (*Admin, error) {
	admin := &Admin{}
	query := `
		SELECT id, user_id, name, created_at
		FROM admins WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
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
	query := `
		SELECT id, user_id, name, created_at
		FROM admins WHERE user_id = $1`

	err := s.db.QueryRow(query, userID).Scan(
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
	query := `
		SELECT id, user_id, name, created_at
		FROM admins 
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
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
	query := `
		UPDATE admins 
		SET name = $2
		WHERE id = $1`

	_, err := s.db.Exec(query, admin.ID, admin.Name)
	return err
}

// Delete deletes an admin
func (s *AdminService) Delete(id uuid.UUID) error {
	query := `DELETE FROM admins WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
