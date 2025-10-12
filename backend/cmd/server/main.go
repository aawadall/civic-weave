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
	var projectService *models.ProjectService
	var applicationService *models.ApplicationService
	var oauthAccountService *models.OAuthAccountService
	var skillClaimService *models.SkillClaimService
	var roleService *models.RoleService
	var volunteerRatingService *models.VolunteerRatingService
	var campaignService *models.CampaignService

	if db != nil {
		userService = models.NewUserService(db)
		volunteerService = models.NewVolunteerService(db)
		adminService = models.NewAdminService(db)
		projectService = models.NewProjectService(db)
		applicationService = models.NewApplicationService(db)
		oauthAccountService = models.NewOAuthAccountService(db)
		skillClaimService = models.NewSkillClaimService(db)
		roleService = models.NewRoleService(db)
		volunteerRatingService = models.NewVolunteerRatingService(db)
		campaignService = models.NewCampaignService(db)
		_ = models.NewEmailVerificationTokenService(db) // for future use
		_ = models.NewPasswordResetTokenService(db)     // for future use
	}

	// Initialize utility services
	emailService := services.NewEmailService(&cfg.Mailgun)
	geocodingService := utils.NewGeocodingService(cfg.Geocoding.NominatimBaseURL)
	embeddingService := services.NewEmbeddingService(cfg.OpenAI.APIKey, cfg.OpenAI.EmbeddingModel)

	// Initialize handlers (only if services are available)
	var authHandler *handlers.AuthHandler
	var googleOAuthHandler *handlers.GoogleOAuthHandler
	var volunteerHandler *handlers.VolunteerHandler
	var projectHandler *handlers.ProjectHandler
	var applicationHandler *handlers.ApplicationHandler
	var matchingHandler *handlers.MatchingHandler
	var skillClaimHandler *handlers.SkillClaimHandler
	var roleHandler *handlers.RoleHandler
	var volunteerRatingHandler *handlers.VolunteerRatingHandler
	var campaignHandler *handlers.CampaignHandler
	var skillHandler *handlers.SkillHandler
	var skillMatchingHandler *handlers.SkillMatchingHandler

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
		googleOAuthHandler = handlers.NewGoogleOAuthHandler(
			userService,
			volunteerService,
			adminService,
			oauthAccountService,
			emailService,
			cfg,
		)
	}
	if volunteerService != nil {
		volunteerHandler = handlers.NewVolunteerHandler(volunteerService, cfg)
	}

	// Initialize skill taxonomy service and handler
	var skillTaxonomyService *models.SkillTaxonomyService
	var skillMatchingService *services.SkillMatchingService
	if db != nil {
		skillTaxonomyService = models.NewSkillTaxonomyService(db)
		skillHandler = handlers.NewSkillHandler(skillTaxonomyService)
		skillMatchingService = services.NewSkillMatchingService(db)
		skillMatchingHandler = handlers.NewSkillMatchingHandler(db, skillTaxonomyService, skillMatchingService)
	}
	if projectService != nil {
		projectHandler = handlers.NewProjectHandler(projectService, geocodingService, cfg)
	}
	if applicationService != nil {
		applicationHandler = handlers.NewApplicationHandler(applicationService, cfg)
	}
	if volunteerService != nil && projectService != nil {
		matchingService := services.NewMatchingService(volunteerService, projectService)
		matchingHandler = handlers.NewMatchingHandler(matchingService, volunteerService, projectService, cfg)
	}

	// Initialize vector-based services and handlers
	var vectorAggregationService *services.VectorAggregationService
	var vectorMatchingService *services.VectorMatchingService
	var adminProfileHandler *handlers.AdminProfileHandler

	if skillClaimService != nil {
		vectorAggregationService = services.NewVectorAggregationService(db, skillClaimService)
		if projectService != nil {
			vectorMatchingService = services.NewVectorMatchingService(db, skillClaimService, vectorAggregationService)
		}
		skillClaimHandler = handlers.NewSkillClaimHandler(
			skillClaimService,
			vectorAggregationService,
			vectorMatchingService,
			embeddingService,
			cfg,
			volunteerService,
		)
	}

	// Initialize new handlers
	if roleService != nil {
		roleHandler = handlers.NewRoleHandler(roleService, cfg)
	}
	if volunteerRatingService != nil && volunteerService != nil {
		volunteerRatingHandler = handlers.NewVolunteerRatingHandler(volunteerRatingService, volunteerService, cfg)
	}
	if campaignService != nil {
		campaignHandler = handlers.NewCampaignHandler(campaignService, emailService, cfg)
	}

	// Initialize admin profile handler
	// var adminSetupHandler *handlers.AdminSetupHandler  // Disabled for security
	if db != nil {
		adminProfileHandler = handlers.NewAdminProfileHandler(db)
		// adminSetupHandler = handlers.NewAdminSetupHandler(userService, adminService, emailService)  // Disabled for security
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
				auth.POST("/register", middleware.RegistrationRateLimiter(), authHandler.Register)
				auth.POST("/login", middleware.LoginRateLimiter(), authHandler.Login)
				auth.POST("/verify-email", authHandler.VerifyEmail)
				// auth.POST("/forgot-password", authHandler.ForgotPassword)
				// auth.POST("/reset-password", authHandler.ResetPassword)
			}
			if googleOAuthHandler != nil {
				auth.POST("/google", middleware.LoginRateLimiter(), googleOAuthHandler.GoogleAuth)
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

			// Volunteer rating routes
			protected.POST("/volunteers/:id/ratings", volunteerRatingHandler.CreateRating)
			protected.GET("/volunteers/:id/scorecard", volunteerRatingHandler.GetVolunteerScorecard)
			protected.GET("/volunteers/:id/ratings", volunteerRatingHandler.ListRatingsForVolunteer)
			protected.GET("/volunteers/top-rated", volunteerRatingHandler.GetTopRatedVolunteers)
			protected.GET("/ratings/my-ratings", volunteerRatingHandler.ListRatingsByRater)
			protected.PUT("/ratings/:id", volunteerRatingHandler.UpdateRating)
			protected.DELETE("/ratings/:id", volunteerRatingHandler.DeleteRating)

			// Project routes (renamed from initiatives)
			protected.GET("/projects", projectHandler.ListProjects)
			protected.POST("/projects", middleware.RequireAnyRole("team_lead", "admin"), projectHandler.CreateProject)
			protected.GET("/projects/:id", projectHandler.GetProject)
			protected.GET("/projects/:id/details", projectHandler.GetProjectWithDetails)
			protected.PUT("/projects/:id", middleware.RequireAnyRole("team_lead", "admin"), projectHandler.UpdateProject)
			protected.DELETE("/projects/:id", middleware.RequireRole("admin"), projectHandler.DeleteProject)

			// Project team management routes
			protected.GET("/projects/:id/signups", projectHandler.GetProjectSignups)
			protected.GET("/projects/:id/team-members", projectHandler.GetProjectTeamMembers)
			protected.POST("/projects/:id/team-members", projectHandler.AddTeamMember)
			protected.PUT("/projects/:id/team-members/:volunteerId", projectHandler.UpdateTeamMemberStatus)
			protected.PUT("/projects/:id/team-lead", middleware.RequireRole("admin"), projectHandler.AssignTeamLead)

			// Application routes
			protected.GET("/applications", applicationHandler.ListApplications)
			protected.POST("/applications", applicationHandler.CreateApplication)
			protected.GET("/applications/:id", applicationHandler.GetApplication)
			protected.PUT("/applications/:id", applicationHandler.UpdateApplication)
			protected.DELETE("/applications/:id", applicationHandler.DeleteApplication)

			// New matching routes (sparse vector system)
			if skillMatchingHandler != nil {
				protected.GET("/matching/my-matches", skillMatchingHandler.GetMyMatches)
				protected.GET("/initiatives/:id/candidate-volunteers", skillMatchingHandler.GetCandidateVolunteers)
				protected.GET("/volunteers/me/recommended-initiatives", skillMatchingHandler.GetRecommendedInitiatives)
				protected.GET("/matching/explanation/:volunteerId/:initiativeId", skillMatchingHandler.GetMatchExplanation)
			}

			// Legacy matching routes (updated to use projects)
			if matchingHandler != nil {
				protected.GET("/matching/legacy/my-matches", matchingHandler.GetMyMatches)
				protected.GET("/matching/legacy/volunteer/:id", matchingHandler.GetMatchesForVolunteer)
				protected.GET("/matching/legacy/project/:id", matchingHandler.GetMatchesForProject)
				protected.GET("/matching/legacy/explanation/:volunteerId/:projectId", matchingHandler.GetMatchExplanation)
			}

			// Skill taxonomy routes (new sparse vector system)
			if skillHandler != nil {
				// Public skill taxonomy
				api.GET("/skills/taxonomy", skillHandler.GetTaxonomy)
				api.POST("/skills/taxonomy", skillHandler.AddSkill)

				// Volunteer skill management
				protected.GET("/volunteers/me/skills", skillHandler.GetVolunteerSkills)
				protected.PUT("/volunteers/me/skills", skillHandler.UpdateVolunteerSkills)
				protected.POST("/volunteers/me/skills", skillHandler.AddVolunteerSkills)
				protected.DELETE("/volunteers/me/skills/:skill_id", skillHandler.RemoveVolunteerSkill)
				protected.GET("/volunteers/me/profile-completion", skillHandler.GetProfileCompletion)

				// Initiative skill management
				protected.GET("/initiatives/:id/skills", skillHandler.GetInitiativeSkills)
				protected.PUT("/initiatives/:id/skills", skillHandler.UpdateInitiativeSkills)
			}

			// Skill claim routes (legacy - keeping for backward compatibility)
			if skillClaimHandler != nil {
				// Volunteer skill management
				protected.GET("/volunteers/me/skill-claims", skillClaimHandler.GetMySkillClaims)
				protected.POST("/volunteers/me/skill-claims", skillClaimHandler.CreateSkillClaim)
				protected.DELETE("/volunteers/me/skill-claims/:id", skillClaimHandler.DeleteSkillClaim)
				protected.GET("/volunteers/me/skills-visibility", skillClaimHandler.GetSkillsVisibility)
				protected.PUT("/volunteers/me/skills-visibility", skillClaimHandler.UpdateSkillsVisibility)
				protected.GET("/volunteers/me/matches", skillClaimHandler.GetTopMatches)
				protected.GET("/volunteers/me/matches/:project_id/explanation", skillClaimHandler.GetMatchExplanation) // TODO: Update handler to use projects

				// Admin skill management
				protected.GET("/admin/skill-claims", middleware.RequireRole("admin"), skillClaimHandler.ListAllSkillClaims)
				protected.PATCH("/admin/skill-claims/:id/weight", middleware.RequireRole("admin"), skillClaimHandler.UpdateSkillWeight)
			}
		}

		// Campaign routes
		if campaignHandler != nil {
			protected.GET("/campaigns", campaignHandler.ListCampaigns)
			protected.POST("/campaigns", campaignHandler.CreateCampaign)
			protected.GET("/campaigns/:id", campaignHandler.GetCampaignByID)
			protected.PUT("/campaigns/:id", campaignHandler.UpdateCampaign)
			protected.DELETE("/campaigns/:id", campaignHandler.DeleteCampaign)
			protected.GET("/campaigns/:id/stats", campaignHandler.GetCampaignStats)
			protected.GET("/campaigns/:id/recipients", campaignHandler.GetCampaignRecipients)
			protected.GET("/campaigns/:id/preview", campaignHandler.PreviewCampaign)
			protected.POST("/campaigns/:id/send", campaignHandler.SendCampaign)
		}

		// Admin profile routes
		if adminProfileHandler != nil {
			protected.GET("/admin/profile", middleware.RequireRole("admin"), adminProfileHandler.GetAdminProfile)
			protected.PUT("/admin/profile", middleware.RequireRole("admin"), adminProfileHandler.UpdateAdminProfile)
			protected.GET("/admin/stats", middleware.RequireRole("admin"), adminProfileHandler.GetSystemStats)
			protected.PUT("/admin/change-password", middleware.RequireRole("admin"), adminProfileHandler.ChangePassword)
		}

		// Role management routes (admin only)
		if roleHandler != nil {
			protected.GET("/admin/roles", roleHandler.ListRoles)
			protected.POST("/admin/roles", roleHandler.CreateRole)
			protected.GET("/admin/roles/:id", roleHandler.GetRoleByID)
			protected.PUT("/admin/roles/:id", roleHandler.UpdateRole)
			protected.DELETE("/admin/roles/:id", roleHandler.DeleteRole)
			protected.GET("/admin/roles/:id/users", roleHandler.ListUsersWithRole)

			// User role assignment routes
			protected.GET("/admin/users", roleHandler.ListAllUsers)
			protected.GET("/admin/users/:id/roles", roleHandler.GetUserRoles)
			protected.POST("/admin/users/:id/roles", roleHandler.AssignRoleToUser)
			protected.DELETE("/admin/users/:id/roles/:roleId", roleHandler.RevokeRoleFromUser)
			protected.GET("/admin/users/:id/role-assignments", roleHandler.GetUserRoleAssignments)
		}
	}

	// Admin setup routes (disabled by default for security)
	// Uncomment only if you need to create another admin user manually
	// if adminSetupHandler != nil {
	// 	router.POST("/api/admin/setup", adminSetupHandler.CreateAdmin)
	// }

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
