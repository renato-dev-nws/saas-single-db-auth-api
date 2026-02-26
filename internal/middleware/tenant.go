package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type tenantCacheData struct {
	TenantID string   `json:"tenant_id"`
	Status   string   `json:"status"`
	Features []string `json:"features"`
}

// TenantMiddleware resolves tenant_id from :url_code
// 1. Extract url_code from route param
// 2. Check Redis cache: "tenant:urlcode:{url_code}"
// 3. If cache miss: query PostgreSQL → SET in Redis with 5min TTL
// 4. Inject tenant_id into context: c.Set("tenant_id", tenantID)
// 5. Inject active features: c.Set("features", []string{...})
// 6. Verify tenant is "active"; if not, return 403
func TenantMiddleware(db *pgxpool.Pool, cache *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		urlCode := c.Param("url_code")
		if urlCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url_code is required"})
			c.Abort()
			return
		}

		// Remove leading slash if present
		urlCode = strings.TrimPrefix(urlCode, "/")

		ctx := context.Background()
		cacheKey := fmt.Sprintf("tenant:urlcode:%s", urlCode)

		// Try cache first
		cached, err := cache.Get(ctx, cacheKey).Result()
		if err == nil {
			var data tenantCacheData
			if json.Unmarshal([]byte(cached), &data) == nil {
				if data.Status != "active" {
					c.JSON(http.StatusForbidden, gin.H{"error": "tenant_not_active"})
					c.Abort()
					return
				}
				c.Set("tenant_id", data.TenantID)
				c.Set("features", data.Features)
				c.Next()
				return
			}
		}

		// Cache miss — query DB
		var tenantID, status string
		err = db.QueryRow(ctx,
			`SELECT id, status FROM tenants WHERE url_code = $1 AND deleted_at IS NULL`,
			urlCode,
		).Scan(&tenantID, &status)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "tenant_not_found"})
			c.Abort()
			return
		}

		if status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": "tenant_not_active"})
			c.Abort()
			return
		}

		// Get active features for this tenant's plan
		features := getActiveTenantFeatures(ctx, db, tenantID)

		// Cache it
		data := tenantCacheData{
			TenantID: tenantID,
			Status:   status,
			Features: features,
		}
		if bytes, err := json.Marshal(data); err == nil {
			cache.Set(ctx, cacheKey, string(bytes), 5*time.Minute)
		}

		c.Set("tenant_id", tenantID)
		c.Set("features", features)
		c.Next()
	}
}

// TenantAccessMiddleware ensures the authenticated user's token tenant_id matches
// the URL-derived tenant_id. Must be placed AFTER TenantMiddleware and an auth middleware.
func TenantAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")            // from TenantMiddleware (URL)
		tokenTenantID := c.GetString("token_tenant_id") // from UserAuthMiddleware or AppAuthMiddleware (JWT)

		if tenantID == "" || tokenTenantID == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access_denied"})
			c.Abort()
			return
		}

		if tenantID != tokenTenantID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access_denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getActiveTenantFeatures(ctx context.Context, db *pgxpool.Pool, tenantID string) []string {
	rows, err := db.Query(ctx,
		`SELECT f.slug
		 FROM tenant_plans tp
		 JOIN plan_features pf ON pf.plan_id = tp.plan_id
		 JOIN features f ON f.id = pf.feature_id
		 WHERE tp.tenant_id = $1 AND tp.is_active = true AND f.is_active = true`,
		tenantID,
	)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var features []string
	for rows.Next() {
		var slug string
		if rows.Scan(&slug) == nil {
			features = append(features, slug)
		}
	}
	return features
}
