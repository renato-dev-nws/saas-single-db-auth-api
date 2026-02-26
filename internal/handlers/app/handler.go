package app

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/saas-single-db-api/internal/cache"
	repo "github.com/saas-single-db-api/internal/repository/app"
	svc "github.com/saas-single-db-api/internal/services/app"
	"github.com/saas-single-db-api/internal/storage"
	"github.com/saas-single-db-api/internal/utils"
)

type Handler struct {
	service *svc.Service
	repo    *repo.Repository
	storage storage.Provider
	cache   *cache.RedisClient
}

func NewHandler(s *svc.Service, r *repo.Repository, st storage.Provider, c *cache.RedisClient) *Handler {
	return &Handler{service: s, repo: r, storage: st, cache: c}
}

// ==================== AUTH ====================

func (h *Handler) Register(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	urlCode := c.Param("url_code")

	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Register(c.Request.Context(), tenantID, urlCode, req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id": result.UserID,
		"token":   result.Token,
	})
}

func (h *Handler) Login(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	urlCode := c.Param("url_code")

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Login(c.Request.Context(), tenantID, urlCode, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   result.Token,
		"user_id": result.UserID,
		"name":    result.Name,
		"email":   result.Email,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token != "" {
		h.cache.SetBlacklist(c.Request.Context(), token)
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *Handler) Me(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("app_user_id")

	result, err := h.service.GetMe(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("app_user_id")

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), tenantID, userID, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = h.service.ForgotPassword(c.Request.Context(), tenantID, req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link was sent"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), tenantID, req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

// ==================== PROFILE ====================

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("app_user_id")
	profile, err := h.repo.GetAppUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"full_name":  profile.FullName,
		"phone":      profile.Phone,
		"document":   profile.Document,
		"avatar_url": profile.AvatarURL,
		"address":    profile.Address,
		"notes":      profile.Notes,
	})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("app_user_id")
	var req struct {
		FullName *string `json:"full_name"`
		Phone    *string `json:"phone"`
		Document *string `json:"document"`
		Address  *string `json:"address"`
		Notes    *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateAppUserProfile(c.Request.Context(), userID, req.FullName, req.Phone, req.Document, req.Address, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("app_user_id")
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	publicURL, storagePath, err := h.storage.Upload(file, header, "app-avatars")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload"})
		return
	}

	_ = h.repo.UpdateAppUserProfileAvatar(c.Request.Context(), userID, publicURL)
	c.JSON(http.StatusOK, gin.H{"path": storagePath, "public_url": publicURL})
}

// ==================== CATALOG (Public) ====================

func (h *Handler) ListProducts(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	pag := utils.GetPagination(c)

	products, total, err := h.repo.ListActiveProducts(c.Request.Context(), tenantID, pag.PageSize, pag.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      products,
		"total":     total,
		"page":      pag.Page,
		"page_size": pag.PageSize,
	})
}

func (h *Handler) GetProduct(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	productID := c.Param("id")
	product, err := h.repo.GetActiveProduct(c.Request.Context(), tenantID, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *Handler) ListServices(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	pag := utils.GetPagination(c)

	services, total, err := h.repo.ListActiveServices(c.Request.Context(), tenantID, pag.PageSize, pag.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list services"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      services,
		"total":     total,
		"page":      pag.Page,
		"page_size": pag.PageSize,
	})
}

func (h *Handler) GetServiceDetail(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	serviceID := c.Param("id")
	service, err := h.repo.GetActiveService(c.Request.Context(), tenantID, serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	c.JSON(http.StatusOK, service)
}
