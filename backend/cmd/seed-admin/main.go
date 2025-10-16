package main

import (
	"fmt"
	"log"
	"os"

	"civicweave/backend/config"
	"civicweave/backend/database"
	"civicweave/backend/models"
	"civicweave/backend/utils"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Get admin credentials from environment variables
	email := getEnv("ADMIN_EMAIL", "admin@civicweave.com")
	password := getEnv("ADMIN_PASSWORD", "")
	name := getEnv("ADMIN_NAME", "System Administrator")

	// Require password to be set via environment variable
	if password == "" {
		log.Fatal("ADMIN_PASSWORD environment variable is required")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create user
	userID := uuid.New()
	user := &models.User{
		ID:            userID,
		Email:         email,
		PasswordHash:  hashedPassword,
		EmailVerified: true, // Admin is pre-verified
	}

	// Insert user
	userService := models.NewUserService(db)
	if err := userService.Create(user); err != nil {
		log.Printf("User might already exist: %v", err)
	}

	// Assign admin role
	roleService := models.NewRoleService(db)
	adminRole, err := roleService.GetByName("admin")
	if err != nil {
		log.Fatal("Failed to get admin role:", err)
	}
	if err := roleService.AssignRoleToUser(user.ID, adminRole.ID, nil); err != nil {
		log.Printf("Failed to assign admin role: %v", err)
	}

	// Create admin profile
	adminID := uuid.New()
	admin := &models.Admin{
		ID:     adminID,
		UserID: userID,
		Name:   name,
	}

	adminService := models.NewAdminService(db)
	if err := adminService.Create(admin); err != nil {
		log.Printf("Admin profile might already exist: %v", err)
	}

	fmt.Printf("✅ Admin user created successfully!\n")
	fmt.Printf("📧 Email: %s\n", email)
	fmt.Printf("🔑 Password: %s\n", password)
	fmt.Printf("👤 Name: %s\n", name)
	fmt.Printf("🆔 User ID: %s\n", userID)
	fmt.Printf("🆔 Admin ID: %s\n", adminID)
	fmt.Printf("\n🌐 Login at: https://civicweave-frontend-peedoie7va-uc.a.run.app/login\n")
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
