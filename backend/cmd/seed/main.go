package main

import (
	"log"

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

	// Create initial admin user
	adminEmail := "admin@civicweave.com"
	adminPassword := "admin123" // Change this in production!

	// Check if admin already exists
	existingUser, _ := userService.GetByEmail(adminEmail)
	if existingUser != nil {
		log.Println("Admin user already exists, skipping seed")
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
		Role:          "admin",
	}

	if err := userService.Create(user); err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	// Create admin profile
	admin := &models.Admin{
		UserID: user.ID,
		Name:   "System Administrator",
	}

	if err := adminService.Create(admin); err != nil {
		log.Fatal("Failed to create admin profile:", err)
	}

	log.Printf("✅ Seeded admin user: %s / %s", adminEmail, adminPassword)
	log.Println("⚠️  IMPORTANT: Change the admin password after first login!")
}
