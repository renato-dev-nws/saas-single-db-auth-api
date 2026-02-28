package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/saas-single-db-api/internal/cache"
	"github.com/saas-single-db-api/internal/config"
	"github.com/saas-single-db-api/internal/database"
	"github.com/saas-single-db-api/internal/email"
	tenantHandler "github.com/saas-single-db-api/internal/handlers/tenant"
	"github.com/saas-single-db-api/internal/middleware"
	tenantRepo "github.com/saas-single-db-api/internal/repository/tenant"
	tenantSvc "github.com/saas-single-db-api/internal/services/tenant"
	"github.com/saas-single-db-api/internal/storage"

	_ "github.com/saas-single-db-api/docs/tenant"
)

// @title Tenant API
// @version 1.0
// @description API de backoffice para tenants do sistema SaaS multi-tenant. Gerencia assinatura, autenticação, membros, produtos, serviços, configurações e muito mais.

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token de autenticação do usuário. Formato: Bearer {token}

func main() {
	cfg := config.Load()

	db := database.NewPostgresPool(cfg.DatabaseURL)
	defer db.Close()

	redisClient := cache.NewRedisClient(cfg.RedisURL)
	defer redisClient.Close()

	storageProvider, err := storage.NewProvider(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage provider: %v", err)
	}

	// Repositories
	repo := tenantRepo.NewRepository(db)

	// Email service
	emailSvc := email.NewService(email.Config{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Password: cfg.SMTPPassword,
		From:     cfg.SMTPFrom,
		AppName:  cfg.AppName,
		BaseURL:  cfg.AppBaseURL,
	}, db)

	// Services
	service := tenantSvc.NewService(repo, redisClient, emailSvc, cfg.JWTSecret, cfg.JWTExpiryHours)

	// Handlers
	handler := tenantHandler.NewHandler(service, repo, storageProvider, redisClient, cfg.JWTSecret, cfg.JWTExpiryHours)

	// Router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		// ─── Public Endpoints ─────────────────────────────
		api.POST("/subscription", handler.Subscribe)
		api.GET("/plans", handler.ListPlans)

		// ─── Auth (Public) ────────────────────────────────
		auth := api.Group("/auth")
		{
			auth.POST("/login", handler.Login)
			auth.GET("/verify-email", handler.VerifyEmail)
		}

		// ─── Auth (Protected) ─────────────────────────────
		protectedAuth := api.Group("/auth")
		protectedAuth.Use(middleware.UserAuthMiddleware(cfg.JWTSecret, redisClient.Inner()))
		{
			protectedAuth.POST("/logout", handler.Logout)
			protectedAuth.GET("/me", handler.Me)
			protectedAuth.POST("/switch/:url_code", handler.SelectTenant)
			protectedAuth.POST("/resend-verification", handler.ResendVerification)
		}

		// ─── Profile (Protected) ──────────────────────────
		profile := api.Group("/profile")
		profile.Use(middleware.UserAuthMiddleware(cfg.JWTSecret, redisClient.Inner()))
		{
			profile.GET("", handler.GetProfile)
			profile.PUT("", handler.UpdateProfile)
			profile.PUT("/password", handler.ChangePassword)
			profile.POST("/avatar", handler.UploadAvatar)
		}

		// ─── Tenant-scoped routes (with TenantMiddleware) ─
		tenantScoped := api.Group("/:url_code")
		tenantScoped.Use(
			middleware.TenantMiddleware(db, redisClient.Inner()),
			middleware.UserAuthMiddleware(cfg.JWTSecret, redisClient.Inner()),
			middleware.TenantAccessMiddleware(),
		)
		{
			// Bootstrap
			tenantScoped.GET("/bootstrap", handler.GetBootstrap)

			// Tenant profile
			tenantScoped.GET("/tenant", handler.GetTenantProfile)
			tenantScoped.PUT("/tenant/profile", handler.UpdateTenantProfile)
			tenantScoped.POST("/tenant/logo", handler.UploadLogo)

			// Members
			members := tenantScoped.Group("/members")
			{
				members.GET("", handler.ListMembers)
				members.GET("/can-add", handler.CanAddMember)
				members.POST("", handler.InviteMember)
				members.GET("/:id", handler.GetMember)
				members.PUT("/:id/role", handler.UpdateMemberRole)
				members.DELETE("/:id", handler.RemoveMember)
			}

			// Roles
			roles := tenantScoped.Group("/roles")
			{
				roles.GET("", handler.ListRoles)
				roles.POST("", handler.CreateRole)
				roles.GET("/:id", handler.GetRole)
				roles.PUT("/:id", handler.UpdateRole)
				roles.DELETE("/:id", handler.DeleteRole)
				roles.GET("/:id/permissions", func(c *gin.Context) {
					roleID := c.Param("id")
					perms, err := repo.GetRolePermissions(c.Request.Context(), roleID)
					if err != nil {
						c.JSON(500, gin.H{"error": "failed to list permissions"})
						return
					}
					c.JSON(200, perms)
				})
				roles.POST("/:id/permissions", handler.AssignPermission)
				roles.DELETE("/:id/permissions/:permId", handler.RemovePermission)
			}

			// Products
			products := tenantScoped.Group("/products")
			{
				products.GET("", handler.ListProducts)
				products.POST("", handler.CreateProduct)
				products.GET("/:id", handler.GetProduct)
				products.PUT("/:id", handler.UpdateProduct)
				products.DELETE("/:id", handler.DeleteProduct)
				products.POST("/:id/image", handler.UploadProductImage)
			}

			// Services
			services := tenantScoped.Group("/services")
			{
				services.GET("", handler.ListServices)
				services.POST("", handler.CreateService)
				services.GET("/:id", handler.GetService)
				services.PUT("/:id", handler.UpdateService)
				services.DELETE("/:id", handler.DeleteService)
				services.POST("/:id/image", handler.UploadServiceImage)
			}

			// Settings
			settings := tenantScoped.Group("/settings")
			{
				settings.GET("/layout", handler.GetLayoutSettings)
				settings.PUT("/layout", handler.UpdateLayoutSettings)
				settings.GET("", handler.ListSettings)
				settings.GET("/:category", handler.GetSetting)
				settings.PUT("/:category", handler.UpsertSetting)
			}

			// App Users (managed from backoffice)
			appUsers := tenantScoped.Group("/app-users")
			{
				appUsers.GET("", handler.ListAppUsers)
				appUsers.GET("/:id", handler.GetAppUser)
				appUsers.PUT("/:id/status", handler.UpdateAppUserStatus)
				appUsers.DELETE("/:id", handler.DeleteAppUser)
			}
		}
	}

	// Serve uploaded files
	r.Static("/uploads", cfg.StorageLocalPath)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	fmt.Printf("🚀 Tenant API starting on port %s\n", cfg.TenantAPIPort)
	if err := r.Run(":" + cfg.TenantAPIPort); err != nil {
		log.Fatalf("Failed to start tenant-api: %v", err)
	}
}
