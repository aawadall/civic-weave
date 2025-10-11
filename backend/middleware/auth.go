package middleware

import (
	"net/http"
	"strings"
	"time"

	"civicweave/backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`  // Deprecated: kept for backward compatibility
	Roles  []string  `json:"roles"` // New: multiple roles support
	jwt.RegisteredClaims
}

// AuthRequired middleware checks for valid JWT token
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Check token expiration
		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)   // Deprecated: kept for backward compatibility
		c.Set("user_roles", claims.Roles) // New: multiple roles support

		c.Next()
	}
}

// RequireRole middleware checks if user has required role (deprecated - use RequireAnyRole)
func RequireRole(requiredRole string) gin.HandlerFunc {
	return RequireAnyRole(requiredRole)
}

// RequireAnyRole middleware checks if user has any of the specified roles
func RequireAnyRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User roles not found"})
			c.Abort()
			return
		}

		roles := userRoles.([]string)
		hasRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllRoles middleware checks if user has all of the specified roles
func RequireAllRoles(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User roles not found"})
			c.Abort()
			return
		}

		roles := userRoles.([]string)
		roleMap := make(map[string]bool)
		for _, role := range roles {
			roleMap[role] = true
		}

		for _, requiredRole := range requiredRoles {
			if !roleMap[requiredRole] {
				c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// OptionalAuth middleware checks for JWT token but doesn't require it
func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || claims.ExpiresAt.Time.Before(time.Now()) {
			c.Next()
			return
		}

		// Set user context if token is valid
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)   // Deprecated: kept for backward compatibility
		c.Set("user_roles", claims.Roles) // New: multiple roles support

		c.Next()
	}
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user *models.User, jwtSecret string) (string, error) {
	// Get user roles
	userService := models.NewUserService(nil) // Will be injected properly in handlers
	roles, err := userService.GetUserRoles(user.ID)
	if err != nil {
		// Fallback to single role if roles service not available
		roles = []models.Role{{Name: user.Role}}
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role, // Deprecated: kept for backward compatibility
		Roles:  roleNames, // New: multiple roles support
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "civicweave",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// GetUserFromContext extracts user information from Gin context
func GetUserFromContext(c *gin.Context) (*UserContext, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, false
	}

	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")   // Deprecated
	roles, _ := c.Get("user_roles") // New

	var rolesList []string
	if roles != nil {
		rolesList = roles.([]string)
	}

	return &UserContext{
		ID:    userID.(uuid.UUID),
		Email: email.(string),
		Role:  role.(string), // Deprecated
		Roles: rolesList,     // New
	}, true
}

// UserContext represents user information from JWT context
type UserContext struct {
	ID    uuid.UUID
	Email string
	Role  string   // Deprecated: kept for backward compatibility
	Roles []string // New: multiple roles support
}

// HasRole checks if user has a specific role
func (uc *UserContext) HasRole(roleName string) bool {
	for _, role := range uc.Roles {
		if role == roleName {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (uc *UserContext) HasAnyRole(roleNames ...string) bool {
	for _, requiredRole := range roleNames {
		for _, userRole := range uc.Roles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// HasAllRoles checks if user has all of the specified roles
func (uc *UserContext) HasAllRoles(roleNames ...string) bool {
	roleMap := make(map[string]bool)
	for _, role := range uc.Roles {
		roleMap[role] = true
	}

	for _, requiredRole := range roleNames {
		if !roleMap[requiredRole] {
			return false
		}
	}
	return true
}

// HasRole helper function for Gin context
func HasRole(c *gin.Context, roleName string) bool {
	userCtx, exists := GetUserFromContext(c)
	if !exists {
		return false
	}
	return userCtx.HasRole(roleName)
}
