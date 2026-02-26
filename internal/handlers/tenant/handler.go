package tenant

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/saas-single-db-api/internal/cache"
	repo "github.com/saas-single-db-api/internal/repository/tenant"
	svc "github.com/saas-single-db-api/internal/services/tenant"
	"github.com/saas-single-db-api/internal/storage"
	"github.com/saas-single-db-api/internal/utils"
)

type Handler struct {
	service   *svc.Service
	repo      *repo.Repository
	storage   storage.Provider
	cache     *cache.RedisClient
	jwtSecret string
	jwtExpiry int
}

func NewHandler(s *svc.Service, r *repo.Repository, st storage.Provider, c *cache.RedisClient, jwtSecret string, jwtExpiry int) *Handler {
	return &Handler{service: s, repo: r, storage: st, cache: c, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

// ==================== PUBLIC: Subscription ====================

func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.repo.ListActivePlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list plans"})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		TenantName   string  `json:"tenant_name" binding:"required"`
		Subdomain    string  `json:"subdomain" binding:"required"`
		IsCompany    bool    `json:"is_company"`
		CompanyName  string  `json:"company_name"`
		PlanID       string  `json:"plan_id" binding:"required"`
		BillingCycle string  `json:"billing_cycle" binding:"required"`
		PromotionID  *string `json:"promotion_id"`
		OwnerName    string  `json:"owner_name" binding:"required"`
		OwnerEmail   string  `json:"owner_email" binding:"required,email"`
		OwnerPass    string  `json:"owner_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subdomain := strings.ToLower(strings.ReplaceAll(req.Subdomain, " ", ""))

	result, err := h.service.Subscribe(c.Request.Context(), svc.SubscribeInput{
		TenantName:   req.TenantName,
		Subdomain:    subdomain,
		IsCompany:    req.IsCompany,
		CompanyName:  req.CompanyName,
		PlanID:       req.PlanID,
		BillingCycle: req.BillingCycle,
		PromotionID:  req.PromotionID,
		OwnerName:    req.OwnerName,
		OwnerEmail:   req.OwnerEmail,
		OwnerPass:    req.OwnerPass,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tenant_id": result.TenantID,
		"user_id":   result.UserID,
		"url_code":  result.URLCode,
		"token":     result.Token,
	})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	if err := h.service.VerifyEmail(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email_verified"})
}

func (h *Handler) ResendVerification(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.service.ResendVerification(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification_email_sent"})
}

// ==================== AUTH ====================

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   result.Token,
		"user_id": result.UserID,
		"name":    result.Name,
		"email":   result.Email,
		"tenants": result.Tenants,
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
	userID := c.GetString("user_id")
	result, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if !utils.CheckPassword(req.CurrentPassword, user.HashPass) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current password is incorrect"})
		return
	}

	hash, _ := utils.HashPassword(req.NewPassword)
	if err := h.repo.UpdateUserPassword(c.Request.Context(), userID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}

func (h *Handler) SelectTenant(c *gin.Context) {
	userID := c.GetString("user_id")
	urlCode := c.Param("url_code")

	tenantID, err := h.repo.GetTenantByURLCode(c.Request.Context(), urlCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	if !h.repo.IsMember(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this tenant"})
		return
	}

	_ = h.repo.UpdateUserLastTenant(c.Request.Context(), userID, urlCode)

	token, err := utils.GenerateUserToken(userID, tenantID, h.jwtSecret, h.jwtExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"tenant_id": tenantID,
		"url_code":  urlCode,
	})
}

// ==================== PROFILE ====================

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	profile, err := h.repo.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		FullName *string `json:"full_name"`
		About    *string `json:"about"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.UpdateUserProfile(c.Request.Context(), userID, req.FullName, req.About); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("user_id")
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	publicURL, storagePath, err := h.storage.Upload(file, header, "avatars")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload"})
		return
	}

	_ = h.repo.UpdateUserProfileAvatar(c.Request.Context(), userID, publicURL)

	c.JSON(http.StatusOK, gin.H{
		"path":       storagePath,
		"public_url": publicURL,
	})
}

// ==================== TENANT CONFIG ====================

func (h *Handler) GetConfig(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")

	tenant, err := h.repo.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	plan, _ := h.repo.GetActiveTenantPlan(c.Request.Context(), tenantID)
	features, _ := h.repo.GetTenantFeatures(c.Request.Context(), tenantID)
	perms, _ := h.repo.GetUserPermissions(c.Request.Context(), userID, tenantID)
	memberCount, _ := h.repo.CountTenantMembers(c.Request.Context(), tenantID)
	isOwner := h.service.IsOwner(c.Request.Context(), userID, tenantID)
	profile, _ := h.repo.GetTenantProfile(c.Request.Context(), tenantID)

	result := gin.H{
		"tenant":       tenant,
		"features":     features,
		"permissions":  perms,
		"is_owner":     isOwner,
		"member_count": memberCount,
	}

	if plan != nil {
		result["plan"] = gin.H{
			"id":               plan.PlanID,
			"name":             plan.PlanName,
			"max_users":        plan.MaxUsers,
			"is_multilang":     plan.IsMultilang,
			"billing_cycle":    plan.BillingCycle,
			"contracted_price": plan.ContractedPrice,
			"promo_price":      plan.PromoPrice,
			"promo_expires_at": plan.PromoExpiresAt,
		}
	}
	if profile != nil {
		result["profile"] = gin.H{
			"about":    profile.About,
			"logo_url": profile.LogoURL,
		}
	}

	c.JSON(http.StatusOK, result)
}

// ==================== TENANT PROFILE ====================

func (h *Handler) GetTenantProfile(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	profile, err := h.repo.GetTenantProfile(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdateTenantProfile(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	if !h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can update tenant profile"})
		return
	}

	var req struct {
		About          *string     `json:"about"`
		CustomSettings interface{} `json:"custom_settings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateTenantProfile(c.Request.Context(), tenantID, req.About, req.CustomSettings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tenant profile updated"})
}

func (h *Handler) UploadLogo(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	if !h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can upload logo"})
		return
	}

	file, header, err := c.Request.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	publicURL, storagePath, err := h.storage.Upload(file, header, "logos")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload"})
		return
	}
	_ = h.repo.UpdateTenantLogo(c.Request.Context(), tenantID, publicURL)

	c.JSON(http.StatusOK, gin.H{
		"path":       storagePath,
		"public_url": publicURL,
	})
}

// ==================== MEMBERS ====================

func (h *Handler) ListMembers(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	members, err := h.repo.ListTenantMembers(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list members"})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *Handler) GetMember(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	memberID := c.Param("id")
	member, err := h.repo.GetMember(c.Request.Context(), tenantID, memberID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}
	c.JSON(http.StatusOK, member)
}

func (h *Handler) CanAddMember(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	canAdd, maxUsers, current, err := h.service.CanAddMember(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"can_add":   canAdd,
		"max_users": maxUsers,
		"current":   current,
	})
}

func (h *Handler) InviteMember(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")

	if !h.service.HasPermission(c.Request.Context(), userID, tenantID, "user_m") &&
		!h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUserID, err := h.service.InviteMember(c.Request.Context(), tenantID, req.Email, req.Name, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": newUserID, "message": "member added"})
}

func (h *Handler) UpdateMemberRole(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	memberID := c.Param("id")

	if !h.service.HasPermission(c.Request.Context(), userID, tenantID, "user_m") &&
		!h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateMemberRole(c.Request.Context(), tenantID, memberID, req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member role updated"})
}

func (h *Handler) RemoveMember(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	memberID := c.Param("id")

	if !h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can remove members"})
		return
	}
	if memberID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove yourself"})
		return
	}

	if err := h.repo.RemoveMember(c.Request.Context(), tenantID, memberID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove member"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

// ==================== ROLES ====================

func (h *Handler) ListRoles(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	roles, err := h.repo.ListTenantRoles(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *Handler) GetRole(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	roleID := c.Param("id")
	role, err := h.repo.GetTenantRoleByID(c.Request.Context(), tenantID, roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	perms, _ := h.repo.GetRolePermissions(c.Request.Context(), roleID)
	c.JSON(http.StatusOK, gin.H{
		"role":        role,
		"permissions": perms,
	})
}

func (h *Handler) CreateRole(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	if !h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can manage roles"})
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := utils.Slugify(req.Title)
	role, err := h.repo.CreateTenantRole(c.Request.Context(), tenantID, req.Title, slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create role"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	roleID := c.Param("id")
	if !h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can manage roles"})
		return
	}

	var req struct {
		Title *string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateTenantRole(c.Request.Context(), tenantID, roleID, req.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *Handler) DeleteRole(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	roleID := c.Param("id")
	if !h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can manage roles"})
		return
	}

	if err := h.repo.DeleteTenantRole(c.Request.Context(), tenantID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}

func (h *Handler) AssignPermission(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	roleID := c.Param("id")
	if !h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can manage permissions"})
		return
	}

	var req struct {
		PermissionID string `json:"permission_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.AssignPermissionToRole(c.Request.Context(), roleID, req.PermissionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign permission"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permission assigned"})
}

func (h *Handler) RemovePermission(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	roleID := c.Param("id")
	if !h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can manage permissions"})
		return
	}

	permID := c.Param("permId")
	if err := h.repo.RemovePermissionFromRole(c.Request.Context(), roleID, permID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove permission"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permission removed"})
}

// ==================== PRODUCTS ====================

func (h *Handler) requireFeature(c *gin.Context, slug string) bool {
	features, _ := c.Get("features")
	if feats, ok := features.([]string); ok {
		for _, f := range feats {
			if f == slug {
				return true
			}
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("feature '%s' not available in your plan", slug)})
	return false
}

func (h *Handler) requirePermission(c *gin.Context, permSlug string) bool {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	if h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		return true
	}
	if h.service.HasPermission(c.Request.Context(), userID, tenantID, permSlug) {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	return false
}

func (h *Handler) ListProducts(c *gin.Context) {
	if !h.requireFeature(c, "products") {
		return
	}
	tenantID := c.GetString("tenant_id")
	pag := utils.GetPagination(c)

	products, total, err := h.repo.ListProducts(c.Request.Context(), tenantID, pag.PageSize, pag.Offset)
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
	if !h.requireFeature(c, "products") {
		return
	}
	tenantID := c.GetString("tenant_id")
	productID := c.Param("id")
	product, err := h.repo.GetProduct(c.Request.Context(), tenantID, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	if !h.requireFeature(c, "products") || !h.requirePermission(c, "prod_c") {
		return
	}
	tenantID := c.GetString("tenant_id")
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
		Price       float64 `json:"price" binding:"required,min=0"`
		SKU         *string `json:"sku"`
		Stock       int     `json:"stock"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.repo.CreateProduct(c.Request.Context(), tenantID, req.Name, req.Description, req.Price, req.SKU, req.Stock)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	if !h.requireFeature(c, "products") || !h.requirePermission(c, "prod_u") {
		return
	}
	tenantID := c.GetString("tenant_id")
	productID := c.Param("id")
	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *float64 `json:"price"`
		SKU         *string  `json:"sku"`
		Stock       *int     `json:"stock"`
		IsActive    *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateProduct(c.Request.Context(), tenantID, productID, req.Name, req.Description, req.Price, req.SKU, req.Stock, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	if !h.requireFeature(c, "products") || !h.requirePermission(c, "prod_d") {
		return
	}
	tenantID := c.GetString("tenant_id")
	productID := c.Param("id")
	if err := h.repo.DeleteProduct(c.Request.Context(), tenantID, productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

func (h *Handler) UploadProductImage(c *gin.Context) {
	if !h.requireFeature(c, "products") || !h.requirePermission(c, "prod_u") {
		return
	}
	tenantID := c.GetString("tenant_id")
	productID := c.Param("id")

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	publicURL, storagePath, err := h.storage.Upload(file, header, "products")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload"})
		return
	}
	_ = h.repo.UpdateProductImage(c.Request.Context(), tenantID, productID, publicURL)

	c.JSON(http.StatusOK, gin.H{"path": storagePath, "public_url": publicURL})
}

// ==================== SERVICES ====================

func (h *Handler) ListServices(c *gin.Context) {
	if !h.requireFeature(c, "services") {
		return
	}
	tenantID := c.GetString("tenant_id")
	pag := utils.GetPagination(c)

	services, total, err := h.repo.ListServices(c.Request.Context(), tenantID, pag.PageSize, pag.Offset)
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

func (h *Handler) GetService(c *gin.Context) {
	if !h.requireFeature(c, "services") {
		return
	}
	tenantID := c.GetString("tenant_id")
	serviceID := c.Param("id")
	service, err := h.repo.GetService(c.Request.Context(), tenantID, serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	c.JSON(http.StatusOK, service)
}

func (h *Handler) CreateService(c *gin.Context) {
	if !h.requireFeature(c, "services") || !h.requirePermission(c, "serv_c") {
		return
	}
	tenantID := c.GetString("tenant_id")
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
		Price       float64 `json:"price" binding:"required,min=0"`
		Duration    *int    `json:"duration"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.repo.CreateService(c.Request.Context(), tenantID, req.Name, req.Description, req.Price, req.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create service"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *Handler) UpdateService(c *gin.Context) {
	if !h.requireFeature(c, "services") || !h.requirePermission(c, "serv_u") {
		return
	}
	tenantID := c.GetString("tenant_id")
	serviceID := c.Param("id")
	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *float64 `json:"price"`
		Duration    *int     `json:"duration"`
		IsActive    *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateService(c.Request.Context(), tenantID, serviceID, req.Name, req.Description, req.Price, req.Duration, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "service updated"})
}

func (h *Handler) DeleteService(c *gin.Context) {
	if !h.requireFeature(c, "services") || !h.requirePermission(c, "serv_d") {
		return
	}
	tenantID := c.GetString("tenant_id")
	serviceID := c.Param("id")
	if err := h.repo.DeleteService(c.Request.Context(), tenantID, serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete service"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "service deleted"})
}

func (h *Handler) UploadServiceImage(c *gin.Context) {
	if !h.requireFeature(c, "services") || !h.requirePermission(c, "serv_u") {
		return
	}
	tenantID := c.GetString("tenant_id")
	serviceID := c.Param("id")

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	publicURL, storagePath, err := h.storage.Upload(file, header, "services")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload"})
		return
	}
	_ = h.repo.UpdateServiceImage(c.Request.Context(), tenantID, serviceID, publicURL)
	c.JSON(http.StatusOK, gin.H{"path": storagePath, "public_url": publicURL})
}

// ==================== SETTINGS ====================

func (h *Handler) ListSettings(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	settings, err := h.repo.ListSettings(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list settings"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetSetting(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	category := c.Param("category")
	setting, err := h.repo.GetSetting(c.Request.Context(), tenantID, category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}
	c.JSON(http.StatusOK, setting)
}

func (h *Handler) UpsertSetting(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")

	if !h.service.HasPermission(c.Request.Context(), userID, tenantID, "setg_m") &&
		!h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var req struct {
		Category string      `json:"category" binding:"required"`
		Data     interface{} `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpsertSetting(c.Request.Context(), tenantID, req.Category, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "setting saved"})
}

// ==================== IMAGES ====================

func (h *Handler) ListImages(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	pag := utils.GetPagination(c)

	images, total, err := h.repo.ListImages(c.Request.Context(), tenantID, pag.PageSize, pag.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list images"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      images,
		"total":     total,
		"page":      pag.Page,
		"page_size": pag.PageSize,
	})
}

func (h *Handler) UploadImage(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	entityType := c.PostForm("entity_type")
	entityID := c.PostForm("entity_id")

	publicURL, storagePath, err := h.storage.Upload(file, header, "images")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload"})
		return
	}

	var eType, eID, uBy *string
	if entityType != "" {
		eType = &entityType
	}
	if entityID != "" {
		eID = &entityID
	}
	uBy = &userID

	providerName := "local"
	if h.storage != nil {
		providerName = "local" // default
	}

	id, err := h.repo.CreateImage(c.Request.Context(), tenantID, header.Filename, storagePath, publicURL, header.Size, header.Header.Get("Content-Type"), providerName, eType, eID, uBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         id,
		"path":       storagePath,
		"public_url": publicURL,
	})
}

func (h *Handler) DeleteImage(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	imageID := c.Param("id")

	storagePath, err := h.repo.DeleteImage(c.Request.Context(), tenantID, imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	_ = h.storage.Delete(storagePath)
	c.JSON(http.StatusOK, gin.H{"message": "image deleted"})
}

// ==================== APP USERS (managed from backoffice) ====================

func (h *Handler) ListAppUsers(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	pag := utils.GetPagination(c)

	users, total, err := h.repo.ListAppUsers(c.Request.Context(), tenantID, pag.PageSize, pag.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list app users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      users,
		"total":     total,
		"page":      pag.Page,
		"page_size": pag.PageSize,
	})
}

func (h *Handler) GetAppUser(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	appUserID := c.Param("id")
	user, err := h.repo.GetAppUser(c.Request.Context(), tenantID, appUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateAppUserStatus(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	appUserID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required,oneof=active blocked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.repo.UpdateAppUserStatus(c.Request.Context(), tenantID, appUserID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *Handler) DeleteAppUser(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	appUserID := c.Param("id")
	if err := h.repo.SoftDeleteAppUser(c.Request.Context(), tenantID, appUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete app user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "app user deleted"})
}
