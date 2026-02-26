package admin

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/saas-single-db-api/internal/cache"
	models "github.com/saas-single-db-api/internal/models/admin"
	"github.com/saas-single-db-api/internal/models/shared"
	tenantModels "github.com/saas-single-db-api/internal/models/tenant"
	svc "github.com/saas-single-db-api/internal/services/admin"
	"github.com/saas-single-db-api/internal/utils"
)

type Handler struct {
	service     *svc.Service
	redisClient *redis.Client
}

func NewHandler(service *svc.Service, redisClient *redis.Client) *Handler {
	return &Handler{service: service, redisClient: redisClient}
}

// --- Auth ---

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, admin, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "admin": admin})
}

func (h *Handler) Logout(c *gin.Context) {
	token := c.GetString("token")
	cache.SetBlacklist(h.redisClient, context.Background(), token, 24*time.Hour)
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "logged_out"})
}

func (h *Handler) Me(c *gin.Context) {
	adminID := c.GetString("admin_id")
	result, err := h.service.GetMe(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin_not_found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	adminID := c.GetString("admin_id")
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.service.Repo().GetAdminByID(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin_not_found"})
		return
	}
	if !utils.CheckPassword(req.CurrentPassword, admin.HashPass) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_current_password"})
		return
	}

	hash, _ := utils.HashPassword(req.NewPassword)
	if err := h.service.Repo().UpdateAdminPassword(c.Request.Context(), adminID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_password"})
		return
	}

	c.JSON(http.StatusOK, shared.MessageResponse{Message: "password_updated"})
}

// --- Sys Users ---

func (h *Handler) ListSysUsers(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	p := utils.GetPagination(c)
	admins, total, err := h.service.Repo().ListAdmins(c.Request.Context(), p.PageSize, p.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_admins"})
		return
	}

	c.JSON(http.StatusOK, shared.PaginatedResponse{
		Data: admins, Total: total, Page: p.Page, PageSize: p.PageSize,
	})
}

func (h *Handler) CreateSysUser(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	var req models.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, _ := utils.HashPassword(req.Password)
	admin, err := h.service.Repo().CreateAdmin(c.Request.Context(), req.FullName, req.Email, hash)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email_already_exists"})
		return
	}

	// Create profile
	h.service.Repo().CreateProfile(c.Request.Context(), admin.ID, req.FullName)

	// Assign role if provided
	if req.RoleSlug != "" {
		role, err := h.service.Repo().GetRoleBySlug(c.Request.Context(), req.RoleSlug)
		if err == nil {
			h.service.Repo().AssignRoleToAdmin(c.Request.Context(), admin.ID, role.ID)
		}
	}

	c.JSON(http.StatusCreated, admin)
}

func (h *Handler) GetSysUser(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	admin, err := h.service.Repo().GetAdminByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin_not_found"})
		return
	}

	profile, _ := h.service.Repo().GetProfile(c.Request.Context(), id)
	roles, _ := h.service.Repo().GetAdminRoles(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"id": admin.ID, "email": admin.Email, "name": admin.Name,
		"status": admin.Status, "profile": profile, "roles": roles,
	})
}

func (h *Handler) UpdateSysUser(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	var req models.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpdateAdmin(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update"})
		return
	}

	c.JSON(http.StatusOK, shared.MessageResponse{Message: "updated"})
}

func (h *Handler) DeleteSysUser(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	if err := h.service.Repo().SoftDeleteAdmin(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete"})
		return
	}

	c.JSON(http.StatusOK, shared.MessageResponse{Message: "deleted"})
}

func (h *Handler) GetSysUserProfile(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	profile, err := h.service.Repo().GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile_not_found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdateSysUserProfile(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpsertProfile(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_profile"})
		return
	}

	c.JSON(http.StatusOK, shared.MessageResponse{Message: "profile_updated"})
}

func (h *Handler) GetMyProfile(c *gin.Context) {
	adminID := c.GetString("admin_id")
	profile, err := h.service.Repo().GetProfile(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile_not_found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdateMyProfile(c *gin.Context) {
	adminID := c.GetString("admin_id")
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpsertProfile(c.Request.Context(), adminID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_profile"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "profile_updated"})
}

// --- Roles ---

func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.service.Repo().ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.service.Repo().CreateRole(c.Request.Context(), req.Title, req.Slug, req.Description)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role_already_exists"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *Handler) GetRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	role, err := h.service.Repo().GetRoleByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role_not_found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *Handler) UpdateRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpdateRole(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_role"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "role_updated"})
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.Repo().DeleteRole(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete_role"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "role_deleted"})
}

func (h *Handler) AssignRole(c *gin.Context) {
	id := c.Param("id")
	var req models.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().AssignRoleToAdmin(c.Request.Context(), id, req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_assign_role"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "role_assigned"})
}

func (h *Handler) RemoveRole(c *gin.Context) {
	id := c.Param("id")
	roleID, _ := strconv.Atoi(c.Param("role_id"))
	if err := h.service.Repo().RemoveRoleFromAdmin(c.Request.Context(), id, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_remove_role"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "role_removed"})
}

func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.service.Repo().ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_permissions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

// --- Tenants ---

func (h *Handler) ListTenants(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "view_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	p := utils.GetPagination(c)
	tenants, total, err := h.service.Repo().ListTenants(c.Request.Context(), p.PageSize, p.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_tenants"})
		return
	}

	c.JSON(http.StatusOK, shared.PaginatedResponse{
		Data: tenants, Total: total, Page: p.Page, PageSize: p.PageSize,
	})
}

func (h *Handler) CreateTenant(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	var req tenantModels.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	tx, err := h.service.Repo().BeginTx(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction_failed"})
		return
	}
	defer tx.Rollback(ctx)

	// Create tenant
	tenantID, err := h.service.Repo().CreateTenant(ctx, tx, req.Name, req.URLCode, req.Subdomain, req.IsCompany, req.CompanyName)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "tenant_already_exists"})
		return
	}

	// Create tenant profile
	h.service.Repo().CreateTenantProfile(ctx, tx, tenantID)

	// Get plan price
	plan, err := h.service.Repo().GetPlanByID(ctx, req.PlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan_not_found"})
		return
	}

	billingCycle := req.BillingCycle
	if billingCycle == "" {
		billingCycle = "monthly"
	}

	basePrice := plan.Price
	contractedPrice := plan.Price
	var promoPrice *float64
	var promoExpiresAt interface{}

	// Apply promotion if provided
	if req.PromotionID != nil {
		promo, err := h.service.Repo().GetPromotionByID(ctx, *req.PromotionID)
		if err == nil && promo.IsActive {
			pp := calculatePromoPrice(basePrice, promo.DiscountType, promo.DiscountValue)
			promoPrice = &pp
			expires := time.Now().AddDate(0, promo.DurationMonths, 0)
			promoExpiresAt = expires
		}
	}

	// Create tenant plan
	if err := h.service.Repo().CreateTenantPlan(ctx, tx, tenantID, req.PlanID, billingCycle, basePrice, contractedPrice, req.PromotionID, promoPrice, promoExpiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_plan"})
		return
	}

	// Copy global roles to tenant
	copyGlobalRolesToTenant(ctx, tx, tenantID)

	// Create owner user if email provided
	var ownerInfo interface{}
	if req.OwnerEmail != "" {
		ownerInfo, err = createOwnerForTenant(ctx, tx, tenantID, req.OwnerEmail, req.OwnerFullName, req.OwnerPassword, req.URLCode)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "owner_creation_failed: " + err.Error()})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit_failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tenant": gin.H{"id": tenantID, "name": req.Name, "url_code": req.URLCode, "status": "active"},
		"owner":  ownerInfo,
	})
}

func (h *Handler) GetTenant(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "view_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	tenant, err := h.service.Repo().GetTenantByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant_not_found"})
		return
	}
	c.JSON(http.StatusOK, tenant)
}

func (h *Handler) UpdateTenant(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	var req tenantModels.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpdateTenant(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "tenant_updated"})
}

func (h *Handler) DeleteTenant(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	if err := h.service.Repo().SoftDeleteTenant(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "tenant_deleted"})
}

func (h *Handler) UpdateTenantStatus(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	var req tenantModels.UpdateTenantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpdateTenantStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_status"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "status_updated"})
}

func (h *Handler) ChangeTenantPlan(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_plans") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	tenantID := c.Param("id")
	var req tenantModels.ChangePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	plan, err := h.service.Repo().GetPlanByID(ctx, req.PlanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan_not_found"})
		return
	}

	tx, err := h.service.Repo().BeginTx(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction_failed"})
		return
	}
	defer tx.Rollback(ctx)

	// Deactivate current plan
	h.service.Repo().DeactivateCurrentPlan(ctx, tx, tenantID)

	basePrice := plan.Price
	contractedPrice := plan.Price
	var promoPrice *float64
	var promoExpiresAt interface{}

	if req.PromotionID != nil {
		promo, err := h.service.Repo().GetPromotionByID(ctx, *req.PromotionID)
		if err == nil && promo.IsActive {
			pp := calculatePromoPrice(basePrice, promo.DiscountType, promo.DiscountValue)
			promoPrice = &pp
			expires := time.Now().AddDate(0, promo.DurationMonths, 0)
			promoExpiresAt = expires
		}
	}

	if err := h.service.Repo().CreateTenantPlan(ctx, tx, tenantID, req.PlanID, req.BillingCycle, basePrice, contractedPrice, req.PromotionID, promoPrice, promoExpiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_plan"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit_failed"})
		return
	}

	c.JSON(http.StatusOK, shared.MessageResponse{Message: "plan_changed"})
}

func (h *Handler) GetTenantPlanHistory(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "view_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	tenantID := c.Param("id")
	history, err := h.service.Repo().GetTenantPlanHistory(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_get_history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history})
}

func (h *Handler) GetTenantMembers(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "view_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	tenantID := c.Param("id")
	members, err := h.service.Repo().GetTenantMembers(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_members"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": members})
}

// --- Plans ---

func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.service.Repo().ListPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_plans"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": plans})
}

func (h *Handler) CreatePlan(c *gin.Context) {
	var req tenantModels.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maxUsers := req.MaxUsers
	if maxUsers == 0 {
		maxUsers = 1
	}

	id, err := h.service.Repo().CreatePlan(c.Request.Context(), req.Name, req.Description, req.Price, maxUsers, req.IsMultilang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_plan"})
		return
	}

	// Add features if provided
	for _, fid := range req.FeatureIDs {
		h.service.Repo().AddFeatureToPlan(c.Request.Context(), id, fid)
	}

	plan, _ := h.service.Repo().GetPlanByID(c.Request.Context(), id)
	c.JSON(http.StatusCreated, plan)
}

func (h *Handler) GetPlan(c *gin.Context) {
	id := c.Param("id")
	plan, err := h.service.Repo().GetPlanByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan_not_found"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *Handler) UpdatePlan(c *gin.Context) {
	id := c.Param("id")
	var req tenantModels.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpdatePlan(c.Request.Context(), id, req.Name, req.Description, req.Price, req.MaxUsers, req.IsMultilang, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_plan"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "plan_updated"})
}

func (h *Handler) DeletePlan(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Repo().DeletePlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete_plan"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "plan_deleted"})
}

func (h *Handler) AddFeatureToPlan(c *gin.Context) {
	planID := c.Param("id")
	var req tenantModels.PlanFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Repo().AddFeatureToPlan(c.Request.Context(), planID, req.FeatureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_add_feature"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "feature_added"})
}

func (h *Handler) RemoveFeatureFromPlan(c *gin.Context) {
	planID := c.Param("id")
	featureID := c.Param("feat_id")
	if err := h.service.Repo().RemoveFeatureFromPlan(c.Request.Context(), planID, featureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_remove_feature"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "feature_removed"})
}

// --- Features ---

func (h *Handler) ListFeatures(c *gin.Context) {
	features, err := h.service.Repo().ListFeatures(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_features"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": features})
}

func (h *Handler) CreateFeature(c *gin.Context) {
	var req tenantModels.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Repo().CreateFeature(c.Request.Context(), req.Title, req.Slug, req.Code, req.Description, req.IsActive)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "feature_already_exists"})
		return
	}

	feature, _ := h.service.Repo().GetFeatureByID(c.Request.Context(), id)
	c.JSON(http.StatusCreated, feature)
}

func (h *Handler) GetFeature(c *gin.Context) {
	id := c.Param("id")
	feature, err := h.service.Repo().GetFeatureByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "feature_not_found"})
		return
	}
	c.JSON(http.StatusOK, feature)
}

func (h *Handler) UpdateFeature(c *gin.Context) {
	id := c.Param("id")
	var req tenantModels.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpdateFeature(c.Request.Context(), id, req.Title, req.Description, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_feature"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "feature_updated"})
}

func (h *Handler) DeleteFeature(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Repo().DeleteFeature(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete_feature"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "feature_deleted"})
}

// --- Promotions ---

func (h *Handler) ListPromotions(c *gin.Context) {
	promos, err := h.service.Repo().ListPromotions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_promotions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": promos})
}

func (h *Handler) CreatePromotion(c *gin.Context) {
	var req tenantModels.CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Repo().CreatePromotion(c.Request.Context(), req.Name, req.Description, req.DiscountType, req.DiscountValue, req.DurationMonths, req.ValidFrom, req.ValidUntil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_promotion"})
		return
	}

	promo, _ := h.service.Repo().GetPromotionByID(c.Request.Context(), id)
	c.JSON(http.StatusCreated, promo)
}

func (h *Handler) GetPromotion(c *gin.Context) {
	id := c.Param("id")
	promo, err := h.service.Repo().GetPromotionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "promotion_not_found"})
		return
	}
	c.JSON(http.StatusOK, promo)
}

func (h *Handler) UpdatePromotion(c *gin.Context) {
	id := c.Param("id")
	var req tenantModels.UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Repo().UpdatePromotion(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_promotion"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "promotion_updated"})
}

func (h *Handler) DeletePromotion(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Repo().DeactivatePromotion(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_deactivate_promotion"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "promotion_deactivated"})
}

// --- Helpers ---

func calculatePromoPrice(basePrice float64, discountType string, discountValue float64) float64 {
	switch discountType {
	case "percent":
		return basePrice * (1 - discountValue/100)
	case "fixed":
		result := basePrice - discountValue
		if result < 0 {
			return 0
		}
		return result
	default:
		return basePrice
	}
}

func copyGlobalRolesToTenant(ctx context.Context, tx pgx.Tx, tenantID string) {
	// Copy global template roles (tenant_id IS NULL) to the new tenant
	tx.Exec(ctx,
		`INSERT INTO user_roles (tenant_id, title, slug)
		 SELECT $1, title, slug FROM user_roles WHERE tenant_id IS NULL`,
		tenantID,
	)

	// Copy role permissions from global templates to new tenant roles
	tx.Exec(ctx,
		`INSERT INTO user_role_permissions (role_id, permission_id)
		 SELECT tr.id, grp.permission_id
		 FROM user_roles tr
		 JOIN user_roles gr ON gr.slug = tr.slug AND gr.tenant_id IS NULL
		 JOIN user_role_permissions grp ON grp.role_id = gr.id
		 WHERE tr.tenant_id = $1`,
		tenantID,
	)
}

func createOwnerForTenant(ctx context.Context, tx pgx.Tx, tenantID, email, fullName, password, urlCode string) (interface{}, error) {
	hashPass, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (name, email, hash_pass, last_tenant_url_code)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		fullName, email, hashPass, urlCode,
	).Scan(&userID)
	if err != nil {
		return nil, err
	}

	tx.Exec(ctx,
		`INSERT INTO user_profiles (user_id, full_name) VALUES ($1, $2)`,
		userID, fullName,
	)

	var ownerRoleID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM user_roles WHERE tenant_id = $1 AND slug = 'owner'`,
		tenantID,
	).Scan(&ownerRoleID)
	if err != nil {
		return nil, err
	}

	tx.Exec(ctx,
		`INSERT INTO tenant_members (user_id, tenant_id, role_id, is_owner) VALUES ($1, $2, $3, true)`,
		userID, tenantID, ownerRoleID,
	)

	return map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"name":    fullName,
	}, nil
}
