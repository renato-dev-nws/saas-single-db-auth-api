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
	appHandler "github.com/saas-single-db-api/internal/handlers/app"
	"github.com/saas-single-db-api/internal/middleware"
	appRepo "github.com/saas-single-db-api/internal/repository/app"
	appSvc "github.com/saas-single-db-api/internal/services/app"
	"github.com/saas-single-db-api/internal/storage"

	_ "github.com/saas-single-db-api/docs/app"
)

// @title App API
// @version 1.0
// @description API pública para usuários finais (app users) dos tenants. Gerencia registro, autenticação, perfil e catálogo de produtos/serviços.

// @host localhost:8082
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token de autenticação do app user. Formato: Bearer {token}

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
	repo := appRepo.NewRepository(db)

	// Services
	service := appSvc.NewService(repo, cfg.JWTSecret, cfg.JWTExpiryHours)

	// Handlers
	handler := appHandler.NewHandler(service, repo, storageProvider, redisClient)

	// Router
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Recovery())

	api := r.Group("/api/v1/:url_code")
	api.Use(middleware.TenantMiddleware(db, redisClient.Inner()))
	{
		// ─── Auth (Public) ────────────────────────────────
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/forgot-password", handler.ForgotPassword)
			auth.POST("/reset-password", handler.ResetPassword)
		}

		// ─── Auth (Protected) ─────────────────────────────
		protectedAuth := api.Group("/auth")
		protectedAuth.Use(
			middleware.AppAuthMiddleware(cfg.JWTSecret, redisClient.Inner()),
			middleware.TenantAccessMiddleware(),
		)
		{
			protectedAuth.POST("/logout", handler.Logout)
			protectedAuth.GET("/me", handler.Me)
		}

		// ─── Profile (Protected) ──────────────────────────
		profile := api.Group("/profile")
		profile.Use(
			middleware.AppAuthMiddleware(cfg.JWTSecret, redisClient.Inner()),
			middleware.TenantAccessMiddleware(),
		)
		{
			profile.GET("", handler.GetProfile)
			profile.PUT("", handler.UpdateProfile)
			profile.PUT("/password", handler.ChangePassword)
			profile.POST("/avatar", handler.UploadAvatar)
		}

		// ─── Catalog (Public) ─────────────────────────────
		catalog := api.Group("/catalog")
		{
			catalog.GET("/products", handler.ListProducts)
			catalog.GET("/products/:id", handler.GetProduct)
			catalog.GET("/services", handler.ListServices)
			catalog.GET("/services/:id", handler.GetServiceDetail)
		}
	}

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	fmt.Printf("🚀 App API starting on port %s\n", cfg.AppAPIPort)
	if err := r.Run(":" + cfg.AppAPIPort); err != nil {
		log.Fatalf("Failed to start app-api: %v", err)
	}
}
