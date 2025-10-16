package main

import (
	"log"
	"os"

	"civicweave/backend/config"
	"civicweave/backend/database"
	"civicweave/backend/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize services
	userService := models.NewUserService(db)
	adminService := models.NewAdminService(db)
	roleService := models.NewRoleService(db)

	// Create default roles
	defaultRoles := []models.Role{
		{Name: "admin", Description: "Administrator with full system access"},
		{Name: "volunteer", Description: "Standard volunteer user"},
		{Name: "team_lead", Description: "Manages projects and volunteers"},
		{Name: "campaign_manager", Description: "Manages outreach campaigns"},
	}

	for _, role := range defaultRoles {
		existingRole, _ := roleService.GetByName(role.Name)
		if existingRole == nil {
			if err := roleService.CreateRole(&role); err != nil {
				log.Printf("Failed to create role %s: %v", role.Name, err)
			} else {
				log.Printf("✅ Created role: %s", role.Name)
			}
		} else {
			log.Printf("Role %s already exists, skipping", role.Name)
		}
	}

	// Create initial admin user
	adminEmail := "admin@civicweave.com"

	// Get admin password from .env file (loaded by godotenv)
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("ADMIN_PASSWORD not found in .env file. Please set ADMIN_PASSWORD in your .env file.")
	}

	// Check if admin already exists
	existingUser, _ := userService.GetByEmail(adminEmail)
	if existingUser != nil {
		log.Println("Admin user already exists, skipping user creation")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash admin password:", err)
	}

	// Create admin user
	user := &models.User{
		Email:         adminEmail,
		PasswordHash:  string(hashedPassword),
		EmailVerified: true, // Skip verification for seeded admin
	}

	if err := userService.Create(user); err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	// Assign admin role
	adminRole, err := roleService.GetByName("admin")
	if err != nil {
		log.Fatal("Failed to get admin role:", err)
	}
	if err := roleService.AssignRoleToUser(user.ID, adminRole.ID, nil); err != nil {
		log.Fatal("Failed to assign admin role:", err)
	}

	// Create admin profile
	admin := &models.Admin{
		UserID: user.ID,
		Name:   "System Administrator",
	}

	if err := adminService.Create(admin); err != nil {
		log.Fatal("Failed to create admin profile:", err)
	}

	log.Printf("✅ Seeded admin user: %s", adminEmail)
	log.Println("⚠️  IMPORTANT: Change the admin password after first login!")
}
