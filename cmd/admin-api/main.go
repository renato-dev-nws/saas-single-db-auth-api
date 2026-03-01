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
	adminHandler "github.com/saas-single-db-api/internal/handlers/admin"
	"github.com/saas-single-db-api/internal/middleware"
	adminRepo "github.com/saas-single-db-api/internal/repository/admin"
	adminSvc "github.com/saas-single-db-api/internal/services/admin"

	_ "github.com/saas-single-db-api/docs/admin"
)

// @title Admin API
// @version 1.0
// @description API de administração do sistema SaaS multi-tenant. Gerencia tenants, planos, features, promoções, usuários do sistema e permissões.

// @host localhost:8081
// @BasePath /api/v1/admin

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token de autenticação. Formato: Bearer {token}

func main() {
	cfg := config.Load()

	db := database.NewPostgresPool(cfg.DatabaseURL)
	defer db.Close()

	redisClient := cache.NewRedisClient(cfg.RedisURL)
	defer redisClient.Close()

	// Repositories
	repo := adminRepo.NewRepository(db)

	// Services
	service := adminSvc.NewService(repo, cfg.JWTSecret, cfg.JWTExpiryHours)

	// Handlers
	handler := adminHandler.NewHandler(service, redisClient.Inner())

	// Router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1/admin")
	{
		// Auth (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", handler.Login)
		}

		// Auth (protected)
		protected := api.Group("")
		protected.Use(middleware.AdminAuthMiddleware(cfg.JWTSecret, redisClient.Inner()))
		{
			protectedAuth := protected.Group("/auth")
			{
				protectedAuth.POST("/logout", handler.Logout)
				protectedAuth.GET("/me", handler.Me)
				protectedAuth.PUT("/password", handler.ChangePassword)
			}

			// Sys Users
			sysUsers := protected.Group("/sys-users")
			{
				sysUsers.GET("", handler.ListSysUsers)
				sysUsers.POST("", handler.CreateSysUser)
				sysUsers.GET("/profile", handler.GetMyProfile)
				sysUsers.PUT("/profile", handler.UpdateMyProfile)
				sysUsers.GET("/:id", handler.GetSysUser)
				sysUsers.PUT("/:id", handler.UpdateSysUser)
				sysUsers.DELETE("/:id", handler.DeleteSysUser)
				sysUsers.GET("/:id/profile", handler.GetSysUserProfile)
				sysUsers.PUT("/:id/profile", handler.UpdateSysUserProfile)
				sysUsers.POST("/:id/roles", handler.AssignRole)
				sysUsers.DELETE("/:id/roles/:role_id", handler.RemoveRole)
			}

			// Roles
			roles := protected.Group("/roles")
			{
				roles.GET("", handler.ListRoles)
				roles.POST("", handler.CreateRole)
				roles.GET("/:id", handler.GetRole)
				roles.PUT("/:id", handler.UpdateRole)
				roles.DELETE("/:id", handler.DeleteRole)
			}

			// Permissions
			protected.GET("/permissions", handler.ListPermissions)

			// Tenants
			tenants := protected.Group("/tenants")
			{
				tenants.GET("", handler.ListTenants)
				tenants.POST("", handler.CreateTenant)
				tenants.GET("/:id", handler.GetTenant)
				tenants.PUT("/:id", handler.UpdateTenant)
				tenants.DELETE("/:id", handler.DeleteTenant)
				tenants.PUT("/:id/status", handler.UpdateTenantStatus)
				tenants.PUT("/:id/plan", handler.ChangeTenantPlan)
				tenants.GET("/:id/plan-history", handler.GetTenantPlanHistory)
				tenants.GET("/:id/members", handler.GetTenantMembers)
			}

			// Plans
			plans := protected.Group("/plans")
			{
				plans.GET("", handler.ListPlans)
				plans.POST("", handler.CreatePlan)
				plans.GET("/:id", handler.GetPlan)
				plans.PUT("/:id", handler.UpdatePlan)
				plans.DELETE("/:id", handler.DeletePlan)
				plans.POST("/:id/features", handler.AddFeatureToPlan)
				plans.DELETE("/:id/features/:feat_id", handler.RemoveFeatureFromPlan)
			}

			// Features
			features := protected.Group("/features")
			{
				features.GET("", handler.ListFeatures)
				features.POST("", handler.CreateFeature)
				features.GET("/:id", handler.GetFeature)
				features.PUT("/:id", handler.UpdateFeature)
				features.DELETE("/:id", handler.DeleteFeature)
			}

			// Promotions
			promos := protected.Group("/promotions")
			{
				promos.GET("", handler.ListPromotions)
				promos.POST("", handler.CreatePromotion)
				promos.GET("/:id", handler.GetPromotion)
				promos.PUT("/:id", handler.UpdatePromotion)
				promos.DELETE("/:id", handler.DeletePromotion)
			}
		}
	}

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	fmt.Printf("🚀 Admin API starting on port %s\n", cfg.AdminAPIPort)
	if err := r.Run(":" + cfg.AdminAPIPort); err != nil {
		log.Fatalf("Failed to start admin-api: %v", err)
	}
}
