package main

import (
	"log"
	"os"

	"civicweave/backend/config"
	"civicweave/backend/database"
	"civicweave/backend/handlers"
	"civicweave/backend/middleware"
	"civicweave/backend/models"
	"civicweave/backend/services"
	"civicweave/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Println("Starting server without database connection...")
		db = nil
	} else {
		// Run migrations
		if err := database.Migrate(db); err != nil {
			log.Printf("Warning: Failed to run migrations: %v", err)
		}
	}

	// Initialize Redis (for future use)
	_ = database.ConnectRedis(cfg.Redis)

	// Initialize services (only if database is available)
	var userService *models.UserService
	var volunteerService *models.VolunteerService
	var adminService *models.AdminService
	var initiativeService *models.InitiativeService
	var applicationService *models.ApplicationService
	var oauthAccountService *models.OAuthAccountService

	if db != nil {
		userService = models.NewUserService(db)
		volunteerService = models.NewVolunteerService(db)
		adminService = models.NewAdminService(db)
		initiativeService = models.NewInitiativeService(db)
		applicationService = models.NewApplicationService(db)
		oauthAccountService = models.NewOAuthAccountService(db)
		_ = models.NewEmailVerificationTokenService(db) // for future use
		_ = models.NewPasswordResetTokenService(db)     // for future use
	}

	// Initialize utility services
	emailService := services.NewEmailService(&cfg.Mailgun)
	geocodingService := utils.NewGeocodingService(cfg.Geocoding.NominatimBaseURL)

	// Initialize handlers (only if services are available)
	var authHandler *handlers.AuthHandler
	var volunteerHandler *handlers.VolunteerHandler
	var initiativeHandler *handlers.InitiativeHandler
	var applicationHandler *handlers.ApplicationHandler
	var matchingHandler *handlers.MatchingHandler

	if userService != nil && volunteerService != nil && adminService != nil && oauthAccountService != nil {
		authHandler = handlers.NewAuthHandler(
			userService,
			volunteerService,
			adminService,
			oauthAccountService,
			emailService,
			geocodingService,
			cfg,
		)
	}
	if volunteerService != nil {
		volunteerHandler = handlers.NewVolunteerHandler(volunteerService, cfg)
	}
	if initiativeService != nil {
		initiativeHandler = handlers.NewInitiativeHandler(initiativeService, geocodingService, cfg)
	}
	if applicationService != nil {
		applicationHandler = handlers.NewApplicationHandler(applicationService, cfg)
	}
	if volunteerService != nil && initiativeService != nil {
		matchingService := services.NewMatchingService(volunteerService, initiativeService)
		matchingHandler = handlers.NewMatchingHandler(matchingService, volunteerService, initiativeService, cfg)
	}

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS())

	// API routes
	api := router.Group("/api")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			if authHandler != nil {
				auth.POST("/register", authHandler.Register)
				auth.POST("/login", authHandler.Login)
				// auth.POST("/google", authHandler.GoogleAuth)
				auth.POST("/verify-email", authHandler.VerifyEmail)
				// auth.POST("/forgot-password", authHandler.ForgotPassword)
				// auth.POST("/reset-password", authHandler.ResetPassword)
			}
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired(cfg.JWT.Secret))
		{
			// User routes
			protected.GET("/me", authHandler.GetProfile)
			// protected.PUT("/me", authHandler.UpdateProfile)

			// Volunteer routes
			protected.GET("/volunteers", volunteerHandler.ListVolunteers)
			protected.POST("/volunteers", volunteerHandler.CreateVolunteer)
			protected.GET("/volunteers/:id", volunteerHandler.GetVolunteer)
			protected.PUT("/volunteers/:id", volunteerHandler.UpdateVolunteer)

			// Initiative routes
			protected.GET("/initiatives", initiativeHandler.ListInitiatives)
			protected.POST("/initiatives", middleware.RequireRole("admin"), initiativeHandler.CreateInitiative)
			protected.GET("/initiatives/:id", initiativeHandler.GetInitiative)
			protected.PUT("/initiatives/:id", middleware.RequireRole("admin"), initiativeHandler.UpdateInitiative)
			protected.DELETE("/initiatives/:id", middleware.RequireRole("admin"), initiativeHandler.DeleteInitiative)

			// Application routes
			protected.GET("/applications", applicationHandler.ListApplications)
			protected.POST("/applications", applicationHandler.CreateApplication)
			protected.GET("/applications/:id", applicationHandler.GetApplication)
			protected.PUT("/applications/:id", applicationHandler.UpdateApplication)
			protected.DELETE("/applications/:id", applicationHandler.DeleteApplication)

			// Matching routes
			protected.GET("/matching/my-matches", matchingHandler.GetMyMatches)
			protected.GET("/matching/volunteer/:id", matchingHandler.GetMatchesForVolunteer)
			protected.GET("/matching/initiative/:id", matchingHandler.GetMatchesForInitiative)
			protected.GET("/matching/explanation/:volunteerId/:initiativeId", matchingHandler.GetMatchExplanation)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
