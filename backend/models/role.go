package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Role represents a system role
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Permissions []string  `json:"permissions" db:"permissions"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	RoleID            uuid.UUID  `json:"role_id" db:"role_id"`
	AssignedByAdminID *uuid.UUID `json:"assigned_by_admin_id" db:"assigned_by_admin_id"`
	AssignedAt        time.Time  `json:"assigned_at" db:"assigned_at"`
}

// UserRoleWithDetails represents a user with their roles (renamed to avoid conflict)
type UserRoleWithDetails struct {
	User  User   `json:"user"`
	Roles []Role `json:"roles"`
}

// RoleService handles role operations
type RoleService struct {
	db *sql.DB
}

// NewRoleService creates a new role service
func NewRoleService(db *sql.DB) *RoleService {
	return &RoleService{db: db}
}

// ListRoles retrieves all roles
func (s *RoleService) ListRoles() ([]Role, error) {
	query := `SELECT id, name, description, permissions, created_at FROM roles ORDER BY name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var permissionsJSON string
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Parse permissions JSON
		if err := ParseJSONArray(permissionsJSON, &role.Permissions); err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// GetRoleByID retrieves a role by ID
func (s *RoleService) GetRoleByID(id uuid.UUID) (*Role, error) {
	role := &Role{}
	var permissionsJSON string
	query := `SELECT id, name, description, permissions, created_at FROM roles WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse permissions JSON
	if err := ParseJSONArray(permissionsJSON, &role.Permissions); err != nil {
		return nil, err
	}

	return role, nil
}

// GetByName retrieves a role by name (alias for GetRoleByName)
func (s *RoleService) GetByName(name string) (*Role, error) {
	return s.GetRoleByName(name)
}

// GetRoleByName retrieves a role by name
func (s *RoleService) GetRoleByName(name string) (*Role, error) {
	role := &Role{}
	var permissionsJSON string
	query := `SELECT id, name, description, permissions, created_at FROM roles WHERE name = $1`

	err := s.db.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse permissions JSON
	if err := ParseJSONArray(permissionsJSON, &role.Permissions); err != nil {
		return nil, err
	}

	return role, nil
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(role *Role) error {
	query := `
		INSERT INTO roles (id, name, description, permissions)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	role.ID = uuid.New()
	permissionsJSON, err := ToJSONArray(role.Permissions)
	if err != nil {
		return err
	}

	return s.db.QueryRow(query, role.ID, role.Name, role.Description, permissionsJSON).
		Scan(&role.CreatedAt)
}

// UpdateRole updates a role
func (s *RoleService) UpdateRole(role *Role) error {
	query := `
		UPDATE roles 
		SET name = $2, description = $3, permissions = $4
		WHERE id = $1`

	permissionsJSON, err := ToJSONArray(role.Permissions)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(query, role.ID, role.Name, role.Description, permissionsJSON)
	return err
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// GetUserRoles retrieves all roles for a user
func (s *RoleService) GetUserRoles(userID uuid.UUID) ([]Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.permissions, r.created_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var permissionsJSON string
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &permissionsJSON, &role.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Parse permissions JSON
		if err := ParseJSONArray(permissionsJSON, &role.Permissions); err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// AssignRoleToUser assigns a role to a user
func (s *RoleService) AssignRoleToUser(userID, roleID uuid.UUID, assignedByAdminID *uuid.UUID) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_by_admin_id, assigned_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, role_id) DO NOTHING`

	_, err := s.db.Exec(query, userID, roleID, assignedByAdminID)
	return err
}

// RevokeRoleFromUser revokes a role from a user
func (s *RoleService) RevokeRoleFromUser(userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := s.db.Exec(query, userID, roleID)
	return err
}

// HasRole checks if a user has a specific role
func (s *RoleService) HasRole(userID uuid.UUID, roleName string) (bool, error) {
	query := `
		SELECT COUNT(1)
		FROM user_roles ur
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.name = $2`

	var count int
	err := s.db.QueryRow(query, userID, roleName).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasAnyRole checks if a user has any of the specified roles
func (s *RoleService) HasAnyRole(userID uuid.UUID, roleNames ...string) (bool, error) {
	if len(roleNames) == 0 {
		return true, nil
	}

	query := `
		SELECT COUNT(1)
		FROM user_roles ur
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.name = ANY($2)`

	var count int
	err := s.db.QueryRow(query, userID, roleNames).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasAllRoles checks if a user has all of the specified roles
func (s *RoleService) HasAllRoles(userID uuid.UUID, roleNames ...string) (bool, error) {
	if len(roleNames) == 0 {
		return true, nil
	}

	query := `
		SELECT COUNT(1)
		FROM user_roles ur
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.name = ANY($2)`

	var count int
	err := s.db.QueryRow(query, userID, roleNames).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == len(roleNames), nil
}

// GetUsersWithRole retrieves all users who have a specific role
func (s *RoleService) GetUsersWithRole(roleName string) ([]User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.email_verified, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE r.name = $1
		ORDER BY u.email`

	rows, err := s.db.Query(query, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserRoleAssignments retrieves all role assignments for a user with details
func (s *RoleService) GetUserRoleAssignments(userID uuid.UUID) ([]UserRole, error) {
	query := `
		SELECT ur.user_id, ur.role_id, ur.assigned_by_admin_id, ur.assigned_at
		FROM user_roles ur
		WHERE ur.user_id = $1
		ORDER BY ur.assigned_at`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []UserRole
	for rows.Next() {
		var assignment UserRole
		err := rows.Scan(&assignment.UserID, &assignment.RoleID, &assignment.AssignedByAdminID, &assignment.AssignedAt)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}
