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

// maskPassword masks a password for logging
func maskPassword(password string) string {
	if password == "" {
		return "(empty)"
	}
	if len(password) <= 4 {
		return "***"
	}
	return password[:2] + "***" + password[len(password)-2:]
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Log database configuration (without password)
	log.Printf("🔧 Database Configuration:")
	log.Printf("   Host: %s", cfg.Database.Host)
	log.Printf("   Port: %s", cfg.Database.Port)
	log.Printf("   Name: %s", cfg.Database.Name)
	log.Printf("   User: %s", cfg.Database.User)
	log.Printf("   SSLMode: %s", cfg.Database.SSLMode)
	log.Printf("   Password: %s", maskPassword(cfg.Database.Password))

	// Initialize database
	log.Println("🔌 Attempting to connect to database...")
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Printf("❌ CRITICAL ERROR: Failed to connect to database: %v", err)
		log.Println("⚠️  WARNING: Starting server WITHOUT database connection - ALL ROUTES WILL FAIL!")
		log.Printf("🔍 Connection string used: host=%s port=%s dbname=%s user=%s sslmode=%s",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.User, cfg.Database.SSLMode)
		db = nil
	} else {
		log.Println("✅ Successfully connected to database")

		// Check database compatibility with runtime version
		runtimeVersion := "1.0.0" // TODO: Get from build info or config
		log.Printf("🔍 Checking database compatibility with runtime version %s...", runtimeVersion)

		compat, err := database.CheckCompatibility(db, runtimeVersion)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to check database compatibility: %v", err)
		} else {
			if !compat.IsCompatible {
				log.Printf("❌ Database compatibility issue: %s", compat.Message)
				log.Println("⚠️  WARNING: Database version may be incompatible with runtime!")
			} else if compat.Status == "warning" {
				log.Printf("⚠️  Database compatibility warning: %s", compat.Message)
			} else {
				log.Println("✅ Database is compatible with runtime version")
			}
		}

		// Run legacy migrations (backward compatibility)
		log.Println("🔄 Running legacy database migrations...")
		if err := database.Migrate(db); err != nil {
			log.Printf("⚠️  Warning: Failed to run legacy migrations: %v", err)
		} else {
			log.Println("✅ Legacy database migrations completed")
		}

		// Run v2 migrations if available
		log.Println("🔄 Checking for enhanced migrations...")
		options := &database.MigrationHookOptions{
			RuntimeVersion: runtimeVersion,
			Quiet:          true,
		}
		if err := database.AutoMigrate(db, runtimeVersion, options); err != nil {
			log.Printf("⚠️  Warning: Enhanced migrations not available or failed: %v", err)
		} else {
			log.Println("✅ Enhanced migrations completed")
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
	var taskHandler *handlers.TaskHandler
	var messageHandler *handlers.MessageHandler

	log.Println("🔧 Initializing handlers...")
	if userService != nil && volunteerService != nil && adminService != nil && oauthAccountService != nil && roleService != nil {
		authHandler = handlers.NewAuthHandler(
			userService,
			volunteerService,
			adminService,
			oauthAccountService,
			roleService,
			emailService,
			geocodingService,
			cfg,
		)
		googleOAuthHandler = handlers.NewGoogleOAuthHandler(
			userService,
			volunteerService,
			adminService,
			oauthAccountService,
			roleService,
			emailService,
			cfg,
		)
		log.Println("✅ Auth handlers initialized")
	} else {
		log.Println("❌ CRITICAL: Auth handlers NOT initialized (database connection failed)")
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
	if roleService != nil && userService != nil {
		roleHandler = handlers.NewRoleHandler(roleService, userService, cfg)
	}
	if volunteerRatingService != nil && volunteerService != nil {
		volunteerRatingHandler = handlers.NewVolunteerRatingHandler(volunteerRatingService, volunteerService, cfg)
	}
	if campaignService != nil {
		campaignHandler = handlers.NewCampaignHandler(campaignService, emailService, cfg)
	}

	// Initialize admin profile handler
	// var adminSetupHandler *handlers.AdminSetupHandler  // Disabled for security
	var adminUserManagementHandler *handlers.AdminUserManagementHandler
	if db != nil {
		adminProfileHandler = handlers.NewAdminProfileHandler(db)
		// adminSetupHandler = handlers.NewAdminSetupHandler(userService, adminService, emailService)  // Disabled for security
	}
	if userService != nil && volunteerService != nil && adminService != nil && roleService != nil {
		adminUserManagementHandler = handlers.NewAdminUserManagementHandler(
			userService,
			volunteerService,
			adminService,
			roleService,
		)
	}

	// Initialize task and message handlers
	var taskService *models.TaskService
	var messageService *models.MessageService
	var broadcastService *models.BroadcastService
	var resourceService *models.ResourceService
	var userDashboardHandler *handlers.UserDashboardHandler

	if db != nil && projectService != nil && volunteerService != nil {
		taskService = models.NewTaskService(db)
		messageService = models.NewMessageService(db)
		broadcastService = models.NewBroadcastService(db)
		resourceService = models.NewResourceService(db)
		taskHandler = handlers.NewTaskHandler(taskService, projectService, volunteerService, messageService)
		messageHandler = handlers.NewMessageHandler(messageService, projectService, userService)
		userDashboardHandler = handlers.NewUserDashboardHandler(
			projectService,
			taskService,
			messageService,
			broadcastService,
			resourceService,
		)
	}

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// Log handler status before registering routes
	log.Println("📋 Handler Status:")
	log.Printf("   authHandler: %v", authHandler != nil)
	log.Printf("   projectHandler: %v", projectHandler != nil)
	log.Printf("   volunteerHandler: %v", volunteerHandler != nil)
	log.Printf("   messageHandler: %v", messageHandler != nil)
	log.Printf("   adminUserManagementHandler: %v", adminUserManagementHandler != nil)

	// API routes
	log.Println("🛣️  Registering API routes...")
	api := router.Group("/api")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			if authHandler != nil {
				auth.POST("/register", middleware.RegistrationRateLimiter(), authHandler.Register)
				auth.POST("/login", middleware.LoginRateLimiter(), authHandler.Login)
				auth.POST("/verify-email", authHandler.VerifyEmail)
				log.Println("✅ Auth routes registered")
			} else {
				log.Println("❌ CRITICAL: Auth routes NOT registered (authHandler is nil)")
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
			protected.PUT("/projects/:id/status", projectHandler.TransitionProjectStatus)
			protected.DELETE("/projects/:id", middleware.RequireRole("admin"), projectHandler.DeleteProject)

			// Project team management routes
			protected.GET("/projects/:id/signups", projectHandler.GetProjectSignups)
			protected.GET("/projects/:id/team-members", projectHandler.GetProjectTeamMembers)
			protected.GET("/projects/:id/team-members-with-details", projectHandler.GetProjectTeamMembersWithDetails)
			protected.POST("/projects/:id/team-members", projectHandler.AddTeamMember)
			protected.PUT("/projects/:id/team-members/:volunteerId", projectHandler.UpdateTeamMemberStatus)
			protected.PUT("/projects/:id/team-lead", middleware.RequireRole("admin"), projectHandler.AssignTeamLead)

			// Project logistics routes
			if projectHandler != nil {
				protected.GET("/projects/:id/logistics", projectHandler.GetLogistics)
				protected.PUT("/projects/:id/logistics", projectHandler.UpdateLogistics)
				protected.POST("/projects/:id/approve-volunteer", projectHandler.ApproveVolunteer)
				protected.DELETE("/projects/:id/volunteers/:volunteerId", projectHandler.RemoveVolunteer)
			}

			// Project task routes
			if taskHandler != nil {
				protected.GET("/projects/:id/tasks", taskHandler.ListTasks)
				protected.GET("/projects/:id/tasks/unassigned", taskHandler.ListUnassignedTasks)
				protected.POST("/projects/:id/tasks", taskHandler.CreateTask)
				protected.GET("/tasks/:id", taskHandler.GetTask)
				protected.PUT("/tasks/:id", taskHandler.UpdateTask)
				protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
				protected.POST("/tasks/:id/assign", taskHandler.SelfAssignTask)
				protected.PUT("/tasks/:id/assign", taskHandler.AssignTask)
				protected.POST("/tasks/:id/updates", taskHandler.AddTaskUpdate)

				// Task comments
				protected.POST("/tasks/:id/comments", taskHandler.AddTaskComment)
				protected.GET("/tasks/:id/comments", taskHandler.GetTaskComments)

				// Task time logging
				protected.POST("/tasks/:id/time-logs", taskHandler.LogTaskTime)
				protected.GET("/tasks/:id/time-logs", taskHandler.GetTaskTimeLogs)

				// Task status transitions
				protected.POST("/tasks/:id/start", taskHandler.StartTask)
				protected.POST("/tasks/:id/mark-blocked", taskHandler.MarkTaskBlocked)
				protected.POST("/tasks/:id/request-takeover", taskHandler.RequestTaskTakeover)
				protected.POST("/tasks/:id/mark-done", taskHandler.MarkTaskDone)
			}

			// Project message routes
			if messageHandler != nil {
				protected.GET("/projects/:id/messages", messageHandler.ListMessages)
				protected.GET("/projects/:id/messages/recent", messageHandler.GetRecentMessages)
				protected.GET("/projects/:id/messages/new", messageHandler.GetNewMessages)
				protected.POST("/projects/:id/messages", messageHandler.SendMessage)
				protected.POST("/projects/:id/messages/read-all", messageHandler.MarkAllAsRead)
				protected.GET("/projects/:id/messages/unread-count", messageHandler.GetUnreadCount)
				protected.PUT("/messages/:id", messageHandler.EditMessage)
				protected.DELETE("/messages/:id", messageHandler.DeleteMessage)
				protected.POST("/messages/:id/read", messageHandler.MarkMessageAsRead)
				protected.GET("/messages/unread-counts", messageHandler.GetAllUnreadCounts)
			}

			// Universal messaging routes
			if messageHandler != nil {
				protected.POST("/messages", messageHandler.SendUniversalMessage)
				protected.GET("/messages/inbox", messageHandler.GetInbox)
				protected.GET("/messages/sent", messageHandler.GetSentMessages)
				protected.GET("/messages/conversations", messageHandler.GetConversations)
				protected.GET("/messages/conversations/:id", messageHandler.GetConversation)
				protected.GET("/messages/unread-count", messageHandler.GetUniversalUnreadCount)
				protected.GET("/messages/recipients/search", messageHandler.SearchRecipients)
			}

			// Broadcast routes
			if broadcastService != nil {
				broadcastHandler := handlers.NewBroadcastHandler(broadcastService)
				protected.GET("/broadcasts", broadcastHandler.ListBroadcasts)
				protected.GET("/broadcasts/:id", broadcastHandler.GetBroadcast)
				protected.POST("/broadcasts", middleware.RequireRole("admin"), broadcastHandler.CreateBroadcast)
				protected.PUT("/broadcasts/:id", broadcastHandler.UpdateBroadcast)
				protected.DELETE("/broadcasts/:id", broadcastHandler.DeleteBroadcast)
				protected.POST("/broadcasts/:id/read", broadcastHandler.MarkBroadcastAsRead)
				protected.GET("/broadcasts/stats", broadcastHandler.GetBroadcastStats)
			}

			// Resource library routes
			if resourceService != nil {
				resourceHandler := handlers.NewResourceHandler(resourceService)
				protected.GET("/resources", resourceHandler.ListResources)
				protected.GET("/resources/:id", resourceHandler.GetResource)
				protected.POST("/resources", middleware.RequireAnyRole("team_lead", "admin"), resourceHandler.CreateResource)
				protected.PUT("/resources/:id", resourceHandler.UpdateResource)
				protected.DELETE("/resources/:id", resourceHandler.DeleteResource)
				protected.GET("/resources/:id/download", resourceHandler.DownloadResource)
				protected.GET("/resources/stats", middleware.RequireRole("admin"), resourceHandler.GetResourceStats)
				protected.GET("/resources/recent", resourceHandler.GetRecentResources)
			}

			// User dashboard routes
			if userDashboardHandler != nil {
				protected.GET("/users/me/projects", userDashboardHandler.GetUserProjects)
				protected.GET("/users/me/tasks", userDashboardHandler.GetUserTasks)
				protected.GET("/users/me/dashboard", userDashboardHandler.GetDashboardData)
			}

			// Application routes
			protected.GET("/applications", applicationHandler.ListApplications)
			protected.POST("/applications", applicationHandler.CreateApplication)
			protected.GET("/applications/:id", applicationHandler.GetApplication)
			protected.PUT("/applications/:id", applicationHandler.UpdateApplication)
			protected.DELETE("/applications/:id", applicationHandler.DeleteApplication)

			// New matching routes (sparse vector system)
			if skillMatchingHandler != nil {
				protected.GET("/matching/my-matches", skillMatchingHandler.GetMyMatches)
				protected.GET("/projects/:id/candidate-volunteers", skillMatchingHandler.GetCandidateVolunteers)
				protected.GET("/volunteers/me/recommended-projects", skillMatchingHandler.GetRecommendedInitiatives)
				protected.GET("/matching/explanation/:volunteerId/:projectId", skillMatchingHandler.GetMatchExplanation)
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

				// Project skill management (legacy initiative endpoints for backward compatibility)
				protected.GET("/initiatives/:id/skills", skillHandler.GetInitiativeSkills)
				protected.PUT("/initiatives/:id/skills", skillHandler.UpdateInitiativeSkills)

				// Project skill management
				protected.GET("/projects/:id/skills", skillHandler.GetProjectSkills)
				protected.PUT("/projects/:id/skills", skillHandler.UpdateProjectSkills)
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

		// Admin user management routes (admin only) - must come before role management to avoid conflicts
		if adminUserManagementHandler != nil {
			protected.GET("/admin/users/:id", middleware.RequireRole("admin"), adminUserManagementHandler.GetUserDetails)
			protected.DELETE("/admin/users/:id", middleware.RequireRole("admin"), adminUserManagementHandler.DeleteUser)
			protected.PUT("/admin/users/:id/verification", middleware.RequireRole("admin"), adminUserManagementHandler.ForceVerificationStatus)
			protected.PUT("/admin/users/:id/password", middleware.RequireRole("admin"), adminUserManagementHandler.ChangeUserPassword)
			log.Println("✅ Admin user management routes registered")
		} else {
			log.Println("❌ Admin user management routes NOT registered (adminUserManagementHandler is nil)")
		}

		// Role management routes (admin only)
		if roleHandler != nil {
			protected.GET("/admin/roles", middleware.RequireRole("admin"), roleHandler.ListRoles)
			protected.POST("/admin/roles", middleware.RequireRole("admin"), roleHandler.CreateRole)
			protected.GET("/admin/roles/:id", middleware.RequireRole("admin"), roleHandler.GetRoleByID)
			protected.PUT("/admin/roles/:id", middleware.RequireRole("admin"), roleHandler.UpdateRole)
			protected.DELETE("/admin/roles/:id", middleware.RequireRole("admin"), roleHandler.DeleteRole)
			protected.GET("/admin/roles/:id/users", middleware.RequireRole("admin"), roleHandler.ListUsersWithRole)

			// User role assignment routes
			protected.GET("/admin/users", middleware.RequireRole("admin"), roleHandler.ListAllUsers)
			protected.GET("/admin/users/:id/roles", middleware.RequireRole("admin"), roleHandler.GetUserRoles)
			protected.POST("/admin/users/:id/roles", middleware.RequireRole("admin"), roleHandler.AssignRoleToUser)
			protected.DELETE("/admin/users/:id/roles/:roleId", middleware.RequireRole("admin"), roleHandler.RevokeRoleFromUser)
			protected.GET("/admin/users/:id/role-assignments", middleware.RequireRole("admin"), roleHandler.GetUserRoleAssignments)
		}
	}

	// Admin setup routes (disabled by default for security)
	// Uncomment only if you need to create another admin user manually
	// if adminSetupHandler != nil {
	// 	router.POST("/api/admin/setup", adminSetupHandler.CreateAdmin)
	// }

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": GetVersionInfo(),
		})
	})

	// Version endpoint
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": GetVersionInfo(),
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("==========================================")
	if authHandler != nil {
		log.Println("✅ Server ready with FULL functionality")
	} else {
		log.Println("⚠️  Server starting with LIMITED functionality (DATABASE CONNECTION FAILED)")
		log.Println("⚠️  Auth routes will return 404 - FIX DATABASE CONNECTION!")
	}
	log.Printf("🚀 Server starting on port %s", port)
	log.Println("==========================================")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
