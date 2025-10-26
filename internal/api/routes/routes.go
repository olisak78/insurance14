package routes

import (
	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/api/middleware"
	"developer-portal-backend/internal/auth"
	"developer-portal-backend/internal/config"
	"developer-portal-backend/internal/repository"
	"developer-portal-backend/internal/service"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// memberRepoAdapter adapts repository.MemberRepository to auth.MemberRepository
type memberRepoAdapter struct {
	repo *repository.MemberRepository
}

func (a *memberRepoAdapter) GetByEmail(email string) (interface{}, error) {
	return a.repo.GetByEmail(email)
}

// SetupRoutes configures all the routes for the application
func SetupRoutes(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(cfg))

	// Initialize validator
	validator := validator.New()

	// Initialize repositories
	organizationRepo := repository.NewOrganizationRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	componentRepo := repository.NewComponentRepository(db)
	landscapeRepo := repository.NewLandscapeRepository(db)
	componentDeploymentRepo := repository.NewComponentDeploymentRepository(db)

	// Initialize services
	organizationService := service.NewOrganizationService(organizationRepo, validator)
	groupService := service.NewGroupService(groupRepo, organizationRepo, validator)
	memberService := service.NewMemberService(memberRepo, validator)
	teamService := service.NewTeamService(teamRepo, groupRepo, organizationRepo, memberRepo, validator)
	projectService := service.NewProjectService(projectRepo, organizationRepo, validator)
	componentService := service.NewComponentService(componentRepo, organizationRepo, validator)
	landscapeService := service.NewLandscapeService(landscapeRepo, organizationRepo, validator)
	componentDeploymentService := service.NewComponentDeploymentService(componentDeploymentRepo, componentRepo, landscapeRepo, validator)
	ldapService := service.NewLDAPService(cfg)
	jiraService := service.NewJiraService(cfg)
	// Initialize Jira PAT on startup: use fixed-name PAT with machine identifier, delete existing if present, then create a new one
	if err := jiraService.InitializePATOnStartup(); err != nil {
		log.Printf("Warning: Jira PAT initialization failed: %v", err)
	}
	jenkinsService := service.NewJenkinsService()
	sonarService := service.NewSonarService(cfg)
	aicoreService := service.NewAICoreService(memberRepo, teamRepo, groupRepo)

	// Initialize auth configuration and services
	authConfig, err := auth.LoadAuthConfig("config/auth.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load auth config: %v", err)
		// Continue without auth if config fails to load
		authConfig = nil
	}

	var authHandler *auth.AuthHandler
	var authMiddleware *auth.AuthMiddleware
	var authService *auth.AuthService
	if authConfig != nil {
		memberRepoAuth := &memberRepoAdapter{repo: memberRepo}
		authService, err = auth.NewAuthService(authConfig, memberRepoAuth)
		if err != nil {
			log.Printf("Warning: Failed to initialize auth service: %v", err)
		} else {
			authHandler = auth.NewAuthHandler(authService)
			authMiddleware = auth.NewAuthMiddleware(authService)
		}
	}

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db)
	organizationHandler := handlers.NewOrganizationHandler(organizationService)
	groupHandler := handlers.NewGroupHandler(groupService)
	memberHandler := handlers.NewMemberHandler(memberService)
	teamHandler := handlers.NewTeamHandler(teamService)
	projectHandler := handlers.NewProjectHandler(projectService)
	componentHandler := handlers.NewComponentHandler(componentService, teamService)
	landscapeHandler := handlers.NewLandscapeHandler(landscapeService)
	componentDeploymentHandler := handlers.NewComponentDeploymentHandler(componentDeploymentService)
	ldapHandler := handlers.NewLDAPHandler(ldapService)
	jiraHandler := handlers.NewJiraHandler(jiraService)
	jenkinsHandler := handlers.NewJenkinsHandler(jenkinsService)
	sonarHandler := handlers.NewSonarHandler(sonarService)
	githubService := service.NewGitHubService(authService)
	githubHandler := handlers.NewGitHubHandler(githubService)
	aicoreHandler := handlers.NewAICoreHandler(aicoreService, validator)

	// Health check routes
	router.GET("/health", healthHandler.Health)
	router.GET("/health/ready", healthHandler.Ready)
	router.GET("/health/live", healthHandler.Live)

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth routes (Backstage-compatible)
	if authHandler != nil {
		auth := router.Group("/api/auth")
		{
			// Provider-specific auth routes
			providerGroup := auth.Group("/:provider")
			{
				providerGroup.GET("/start", authHandler.Start)
				providerGroup.GET("/handler/frame", authHandler.HandlerFrame)
				providerGroup.GET("/refresh", authHandler.Refresh)
				providerGroup.POST("/logout", authHandler.Logout)
			}

			// Helper endpoint for token validation (not part of Backstage spec)
			auth.POST("/validate", authHandler.ValidateToken)
		}
	}

	// API v1 routes - All endpoints require authentication
	v1 := router.Group("/api/v1")

	// Apply auth middleware to require authentication for all API endpoints
	if authMiddleware != nil {
		v1.Use(authMiddleware.RequireAuth())
	}

	{
		// Organization routes
		organizations := v1.Group("/organizations")
		{
			organizations.GET("", organizationHandler.ListOrganizations)
			organizations.POST("", organizationHandler.CreateOrganization)
			organizations.GET("/:id", organizationHandler.GetOrganization)
			organizations.GET("/by-name/:name", organizationHandler.GetOrganizationByName)
			organizations.PUT("/:id", organizationHandler.UpdateOrganization)
			organizations.DELETE("/:id", organizationHandler.DeleteOrganization)
			organizations.GET("/:id/members", memberHandler.GetMembersByOrganization)
			organizations.GET("/:id/groups", groupHandler.GetGroupsByOrganization)
			organizations.GET("/:id/teams", teamHandler.GetTeamsByOrganization)
			organizations.GET("/:id/projects", projectHandler.GetProjectsByOrganization)
			organizations.GET("/:id/components", componentHandler.GetComponentsByOrganization)
			organizations.GET("/:id/landscapes", landscapeHandler.GetLandscapesByOrganization)
		}

		// Group routes
		groups := v1.Group("/groups")
		{
			groups.GET("", groupHandler.GetGroupsByOrganization) // Requires organization_id parameter
			groups.POST("", groupHandler.CreateGroup)
			groups.GET("/:id", groupHandler.GetGroup)
			groups.PUT("/:id", groupHandler.UpdateGroup)
			groups.DELETE("/:id", groupHandler.DeleteGroup)
			groups.GET("/:id/teams", groupHandler.GetGroupWithTeams)
			groups.GET("/by-name/:name", groupHandler.GetGroupByName) // Requires organization_id parameter
			groups.GET("/search", groupHandler.SearchGroups)          // Requires organization_id and q parameters
		}

		// Member routes
		members := v1.Group("/members")
		{
			members.GET("", memberHandler.GetMembersByOrganization) // Requires organization_id parameter
			members.POST("", memberHandler.CreateMember)
			members.GET("/:id", memberHandler.GetMember)
			members.PUT("/:id", memberHandler.UpdateMember)
			members.DELETE("/:id", memberHandler.DeleteMember)
			members.GET("/:id/quick-links", memberHandler.GetQuickLinks)      // Get quick links for a member
			members.POST("/:id/quick-links", memberHandler.AddQuickLink)      // Add a quick link to a member
			members.DELETE("/:id/quick-links", memberHandler.RemoveQuickLink) // Remove a quick link from a member
		}

		// Team routes
		teams := v1.Group("/teams")
		{
			teams.GET("", teamHandler.GetAllTeams) // Optional organization_id parameter
			teams.GET("/all", teamHandler.GetAllTeamsDeprecated)
			teams.POST("", teamHandler.CreateTeam)
			teams.GET("/:id", teamHandler.GetTeam)
			teams.PUT("/:id", teamHandler.UpdateTeam)
			teams.DELETE("/:id", teamHandler.DeleteTeam)
			teams.GET("/:id/members", teamHandler.GetTeamWithMembers)
			teams.GET("/:id/details", teamHandler.GetTeamWithMembers)
			teams.GET("/:id/components", teamHandler.GetTeamComponents)           // Get components by team ID
			teams.POST("/:id/links", teamHandler.AddLink)                         // Add a link to a team
			teams.DELETE("/:id/links", teamHandler.RemoveLink)                    // Remove a link from a team
			teams.PUT("/:id/links", teamHandler.UpdateLinks)                      // Update all links for a team
			teams.GET("/by-name/:name", teamHandler.GetTeamByName)                // Requires organization_id parameter
			teams.GET("/by-name/:name/members", teamHandler.GetTeamMembersByName) // Requires organization_id parameter
		}

		// Project routes
		projects := v1.Group("/projects")
		{
			projects.GET("", projectHandler.GetProjectsByOrganization) // Requires organization_id parameter
			projects.POST("", projectHandler.CreateProject)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.DeleteProject)
			projects.GET("/:id/organization", projectHandler.GetProjectWithOrganization)
			projects.GET("/:id/components", projectHandler.GetProjectWithComponents)
			projects.GET("/:id/landscapes", projectHandler.GetProjectWithLandscapes)
			projects.GET("/status/:status", projectHandler.GetProjectsByStatus)
		}

		// Component routes
		components := v1.Group("/components")
		{
			components.GET("", componentHandler.ListComponents)
			components.POST("", componentHandler.CreateComponent)
			components.GET("/by-name/:name", componentHandler.GetComponentByName)  // Requires organization_id parameter
			components.GET("/by-team/:id", componentHandler.GetComponentsByTeamID) // Get components by team ID
			components.GET("/:id", componentHandler.GetComponent)
			components.PUT("/:id", componentHandler.UpdateComponent)
			components.DELETE("/:id", componentHandler.DeleteComponent)
			components.GET("/:id/projects", componentHandler.GetComponentWithProjects)
			components.GET("/:id/deployments", componentHandler.GetComponentWithDeployments)
			components.GET("/:id/ownerships", componentHandler.GetComponentWithOwnerships)
			components.GET("/:id/details", componentHandler.GetComponentWithFullDetails)
		}

		// Landscape routes
		landscapes := v1.Group("/landscapes")
		{
			landscapes.GET("", landscapeHandler.ListLandscapes)
			landscapes.POST("", landscapeHandler.CreateLandscape)
			landscapes.GET("/:id", landscapeHandler.GetLandscape)
			landscapes.PUT("/:id", landscapeHandler.UpdateLandscape)
			landscapes.DELETE("/:id", landscapeHandler.DeleteLandscape)
			landscapes.GET("/:id/projects", landscapeHandler.GetLandscapeWithProjects)
			landscapes.GET("/:id/deployments", landscapeHandler.GetLandscapeWithDeployments)
			landscapes.GET("/:id/details", landscapeHandler.GetLandscapeWithFullDetails)
			landscapes.GET("/environment/:environment", landscapeHandler.GetLandscapesByEnvironment)
		}

		// Component Deployment routes
		componentDeployments := v1.Group("/component-deployments")
		{
			componentDeployments.GET("", componentDeploymentHandler.ListComponentDeployments)
			componentDeployments.POST("", componentDeploymentHandler.CreateComponentDeployment)
			componentDeployments.GET("/:id", componentDeploymentHandler.GetComponentDeployment)
			componentDeployments.PUT("/:id", componentDeploymentHandler.UpdateComponentDeployment)
			componentDeployments.DELETE("/:id", componentDeploymentHandler.DeleteComponentDeployment)
			componentDeployments.GET("/:id/details", componentDeploymentHandler.GetComponentDeploymentWithDetails)
		}

		// LDAP routes
		ldap := v1.Group("/ldap")
		{
			ldap.GET("/users/search", ldapHandler.UserSearch)
		}

		// Jira routes - Consolidated endpoints
		jira := v1.Group("/jira")
		{
			jira.GET("/issues", jiraHandler.GetIssues)                 // GET /jira/issues?project=SAPBTPCFS&status=Open,In Progress&team=MyTeam
			jira.GET("/issues/me", jiraHandler.GetMyIssues)            // GET /jira/issues/me?status=Open&count_only=true
			jira.GET("/issues/me/count", jiraHandler.GetMyIssuesCount) // GET /jira/issues/me/count?status=Resolved&date=2023-01-01
		}

		// GitHub routes
		github := v1.Group("/github")
		{
			github.GET("/pull-requests", githubHandler.GetMyPullRequests)
			github.GET("/prs", githubHandler.GetMyPullRequests) // Convenient alias
			github.GET("/contributions", githubHandler.GetUserTotalContributions)
		}

		// Sonar routes
		sonar := v1.Group("/sonar")
		{
			sonar.GET("/measures", sonarHandler.GetMeasures)
		}

		// Self-service routes (for Jenkins, and future services like Kubernetes, etc.)
		selfService := v1.Group("/self-service")
		{
			// Jenkins self-service endpoints
			jenkins := selfService.Group("/jenkins")
			{
				jenkins.GET("/:jaasName/:jobName/parameters", jenkinsHandler.GetJobParameters)
				jenkins.POST("/:jaasName/:jobName/trigger", jenkinsHandler.TriggerJob)
			}
		}

		// AI Core routes
		aicore := v1.Group("/ai-core")
		{
			aicore.GET("/deployments", aicoreHandler.GetDeployments)
			aicore.GET("/deployments/:deploymentId", aicoreHandler.GetDeploymentDetails)
			aicore.GET("/models", aicoreHandler.GetModels)
			aicore.GET("/configurations", aicoreHandler.GetConfigurations)
			aicore.POST("/configurations", aicoreHandler.CreateConfiguration)
			aicore.POST("/deployments", aicoreHandler.CreateDeployment)
			aicore.PATCH("/deployments/:deploymentId", aicoreHandler.UpdateDeployment)
			aicore.DELETE("/deployments/:deploymentId", aicoreHandler.DeleteDeployment)
		}

		// Nested resource routes moved to respective groups to avoid conflicts
		// Landscape-specific component deployments route moved to landscapes group
	}

	// Catch-all route for undefined endpoints
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":      "Endpoint not found",
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"request_id": c.GetString("request_id"),
		})
	})

	return router
}

// SetupHealthRoutes sets up only health check routes (useful for testing)
func SetupHealthRoutes(db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	healthHandler := handlers.NewHealthHandler(db)
	router.GET("/health", healthHandler.Health)
	router.GET("/health/ready", healthHandler.Ready)
	router.GET("/health/live", healthHandler.Live)

	return router
}
