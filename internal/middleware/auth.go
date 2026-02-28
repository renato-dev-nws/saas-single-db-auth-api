package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/saas-single-db-api/internal/cache"
	"github.com/saas-single-db-api/internal/utils"
)

// AdminAuthMiddleware validates JWT for saas_admin_users
func AdminAuthMiddleware(jwtSecret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization_required"})
			c.Abort()
			return
		}

		// Check blacklist
		if cache.IsBlacklisted(redisClient, context.Background(), token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token_invalidated"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateAdminToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			c.Abort()
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("token", token)
		c.Next()
	}
}

// UserAuthMiddleware validates JWT for backoffice users
func UserAuthMiddleware(jwtSecret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization_required"})
			c.Abort()
			return
		}

		// Check blacklist
		if cache.IsBlacklisted(redisClient, context.Background(), token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token_invalidated"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateUserToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("token_tenant_id", claims.TenantID)
		c.Set("token", token)
		c.Next()
	}
}

// AppAuthMiddleware validates JWT for tenant_app_users
func AppAuthMiddleware(jwtSecret string, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization_required"})
			c.Abort()
			return
		}

		// Check blacklist
		if cache.IsBlacklisted(redisClient, context.Background(), token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token_invalidated"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateAppUserToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			c.Abort()
			return
		}

		c.Set("app_user_id", claims.AppUserID)
		c.Set("token_tenant_id", claims.TenantID)
		c.Set("token", token)
		c.Next()
	}
}

// extractToken extracts the Bearer token from the Authorization header
func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
