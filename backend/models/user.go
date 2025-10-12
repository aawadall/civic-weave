package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system (unified auth)
type User struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	Role          string    `json:"role" db:"role"` // Deprecated: kept for backward compatibility
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// UserWithRoles represents a user with their roles
type UserWithRoles struct {
	User
	Roles []Role `json:"roles"`
}

// UserService handles user operations
type UserService struct {
	db *sql.DB
}

// NewUserService creates a new user service
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// Create creates a new user
func (s *UserService) Create(user *User) error {
	// Validate input
	if err := ValidateUser(user); err != nil {
		return err
	}

	// Sanitize input
	user.Email = SanitizeString(user.Email)

	user.ID = uuid.New()
	return s.db.QueryRow(userCreateQuery, user.ID, user.Email, user.PasswordHash, user.EmailVerified, user.Role).
		Scan(&user.CreatedAt, &user.UpdatedAt)
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(id uuid.UUID) (*User, error) {
	user := &User{}

	err := s.db.QueryRow(userGetByIDQuery, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(email string) (*User, error) {
	user := &User{}

	err := s.db.QueryRow(userGetByEmailQuery, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// Update updates a user
func (s *UserService) Update(user *User) error {
	return s.db.QueryRow(userUpdateQuery, user.ID, user.Email, user.PasswordHash, user.EmailVerified, user.Role).
		Scan(&user.UpdatedAt)
}

// VerifyEmail marks a user's email as verified
func (s *UserService) VerifyEmail(userID uuid.UUID) error {
	_, err := s.db.Exec(userVerifyEmailQuery, userID)
	return err
}

// Delete deletes a user
func (s *UserService) Delete(id uuid.UUID) error {
	_, err := s.db.Exec(userDeleteQuery, id)
	return err
}

// GetUserRoles retrieves all roles for a user
func (s *UserService) GetUserRoles(userID uuid.UUID) ([]Role, error) {
	roleService := NewRoleService(s.db)
	return roleService.GetUserRoles(userID)
}

// GetUserWithRoles retrieves a user with their roles
func (s *UserService) GetUserWithRoles(userID uuid.UUID) (*UserWithRoles, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	roles, err := s.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	return &UserWithRoles{
		User:  *user,
		Roles: roles,
	}, nil
}

// HasRole checks if a user has a specific role
func (s *UserService) HasRole(userID uuid.UUID, roleName string) (bool, error) {
	roleService := NewRoleService(s.db)
	return roleService.HasRole(userID, roleName)
}

// HasAnyRole checks if a user has any of the specified roles
func (s *UserService) HasAnyRole(userID uuid.UUID, roleNames ...string) (bool, error) {
	roleService := NewRoleService(s.db)
	return roleService.HasAnyRole(userID, roleNames...)
}

// ListAllUsers retrieves all users (for admin purposes)
func (s *UserService) ListAllUsers() ([]User, error) {
	rows, err := s.db.Query(userListAllQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
