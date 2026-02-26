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
	_ "github.com/saas-single-db-api/internal/models/swagger"
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

// Login godoc
// @Summary Login de administrador
// @Description Autentica um administrador do sistema com email e senha
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Credenciais de login"
// @Success 200 {object} swagger.AdminLoginResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	token, admin, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "admin": admin})
}

// Logout godoc
// @Summary Logout de administrador
// @Description Invalida o token JWT do administrador
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.MessageResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	token := c.GetString("token")
	cache.SetBlacklist(h.redisClient, context.Background(), token, 24*time.Hour)
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "logged_out"})
}

// Me godoc
// @Summary Dados do administrador autenticado
// @Description Retorna os dados do administrador logado
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.AdminUserDTO
// @Failure 401 {object} swagger.ErrorResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	adminID := c.GetString("admin_id")
	result, err := h.service.GetMe(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "admin_not_found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ChangePassword godoc
// @Summary Alterar senha do administrador
// @Description Altera a senha do administrador autenticado
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ChangePasswordRequest true "Senhas atual e nova"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /auth/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	adminID := c.GetString("admin_id")
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
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

// ListSysUsers godoc
// @Summary Listar administradores do sistema
// @Description Retorna lista paginada de administradores. Requer permissão manage_sys_users.
// @Tags Sys Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Página" default(1)
// @Param page_size query int false "Itens por página" default(20)
// @Success 200 {object} swagger.PaginatedResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /sys-users [get]
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

// CreateSysUser godoc
// @Summary Criar administrador do sistema
// @Description Cria um novo administrador. Requer permissão manage_sys_users.
// @Tags Sys Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateAdminRequest true "Dados do administrador"
// @Success 201 {object} swagger.AdminUserDTO
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 409 {object} swagger.ErrorResponse
// @Router /sys-users [post]
func (h *Handler) CreateSysUser(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	var req models.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
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

// GetSysUser godoc
// @Summary Obter administrador por ID
// @Description Retorna detalhes de um administrador específico. Requer permissão manage_sys_users.
// @Tags Sys Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do administrador"
// @Success 200 {object} swagger.SysUserDetailResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /sys-users/{id} [get]
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

// UpdateSysUser godoc
// @Summary Atualizar administrador
// @Description Atualiza dados de um administrador. Requer permissão manage_sys_users.
// @Tags Sys Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do administrador"
// @Param request body models.UpdateAdminRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /sys-users/{id} [put]
func (h *Handler) UpdateSysUser(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	var req models.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpdateAdmin(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update"})
		return
	}

	c.JSON(http.StatusOK, shared.MessageResponse{Message: "updated"})
}

// DeleteSysUser godoc
// @Summary Remover administrador
// @Description Remove (soft delete) um administrador. Requer permissão manage_sys_users.
// @Tags Sys Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do administrador"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /sys-users/{id} [delete]
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

// GetSysUserProfile godoc
// @Summary Obter perfil de administrador
// @Description Retorna o perfil de um administrador por ID. Requer permissão manage_sys_users.
// @Tags Sys Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do administrador"
// @Success 200 {object} swagger.AdminProfileDTO
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /sys-users/{id}/profile [get]
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

// UpdateSysUserProfile godoc
// @Summary Atualizar perfil de administrador
// @Description Atualiza o perfil de um administrador por ID. Requer permissão manage_sys_users.
// @Tags Sys Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do administrador"
// @Param request body models.UpdateProfileRequest true "Dados do perfil"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /sys-users/{id}/profile [put]
func (h *Handler) UpdateSysUserProfile(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_sys_users") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpsertProfile(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_profile"})
		return
	}

	c.JSON(http.StatusOK, shared.MessageResponse{Message: "profile_updated"})
}

// GetMyProfile godoc
// @Summary Obter meu perfil
// @Description Retorna o perfil do administrador autenticado
// @Tags Sys Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.AdminProfileDTO
// @Failure 401 {object} swagger.ErrorResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /sys-users/profile [get]
func (h *Handler) GetMyProfile(c *gin.Context) {
	adminID := c.GetString("admin_id")
	profile, err := h.service.Repo().GetProfile(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile_not_found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// UpdateMyProfile godoc
// @Summary Atualizar meu perfil
// @Description Atualiza o perfil do administrador autenticado
// @Tags Sys Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateProfileRequest true "Dados do perfil"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /sys-users/profile [put]
func (h *Handler) UpdateMyProfile(c *gin.Context) {
	adminID := c.GetString("admin_id")
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpsertProfile(c.Request.Context(), adminID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_profile"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "profile_updated"})
}

// --- Roles ---

// ListRoles godoc
// @Summary Listar roles do sistema
// @Description Retorna todas as roles de administradores
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.AdminRoleListResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.service.Repo().ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// CreateRole godoc
// @Summary Criar role
// @Description Cria uma nova role de administrador
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateRoleRequest true "Dados da role"
// @Success 201 {object} swagger.AdminRoleDTO
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 409 {object} swagger.ErrorResponse
// @Router /roles [post]
func (h *Handler) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	role, err := h.service.Repo().CreateRole(c.Request.Context(), req.Title, req.Slug, req.Description)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role_already_exists"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// GetRole godoc
// @Summary Obter role por ID
// @Description Retorna uma role específica
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da role"
// @Success 200 {object} swagger.AdminRoleDTO
// @Failure 404 {object} swagger.ErrorResponse
// @Router /roles/{id} [get]
func (h *Handler) GetRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	role, err := h.service.Repo().GetRoleByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role_not_found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// UpdateRole godoc
// @Summary Atualizar role
// @Description Atualiza uma role existente
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da role"
// @Param request body models.UpdateRoleRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /roles/{id} [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpdateRole(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_role"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "role_updated"})
}

// DeleteRole godoc
// @Summary Remover role
// @Description Remove uma role do sistema
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da role"
// @Success 200 {object} swagger.MessageResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /roles/{id} [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.Repo().DeleteRole(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete_role"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "role_deleted"})
}

// AssignRole godoc
// @Summary Atribuir role a administrador
// @Description Atribui uma role a um administrador
// @Tags Sys Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do administrador"
// @Param request body models.AssignRoleRequest true "ID da role"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /sys-users/{id}/roles [post]
func (h *Handler) AssignRole(c *gin.Context) {
	id := c.Param("id")
	var req models.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().AssignRoleToAdmin(c.Request.Context(), id, req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_assign_role"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "role_assigned"})
}

// RemoveRole godoc
// @Summary Remover role de administrador
// @Description Remove a atribuição de uma role de um administrador
// @Tags Sys Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do administrador"
// @Param role_id path int true "ID da role"
// @Success 200 {object} swagger.MessageResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /sys-users/{id}/roles/{role_id} [delete]
func (h *Handler) RemoveRole(c *gin.Context) {
	id := c.Param("id")
	roleID, _ := strconv.Atoi(c.Param("role_id"))
	if err := h.service.Repo().RemoveRoleFromAdmin(c.Request.Context(), id, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_remove_role"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "role_removed"})
}

// ListPermissions godoc
// @Summary Listar permissões
// @Description Retorna todas as permissões do sistema
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.AdminPermissionListResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.service.Repo().ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_permissions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": perms})
}

// --- Tenants ---

// ListTenants godoc
// @Summary Listar tenants
// @Description Retorna lista paginada de tenants. Requer permissão view_tenants.
// @Tags Tenants
// @Produce json
// @Security BearerAuth
// @Param page query int false "Página" default(1)
// @Param page_size query int false "Itens por página" default(20)
// @Success 200 {object} swagger.PaginatedResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /tenants [get]
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

// CreateTenant godoc
// @Summary Criar tenant
// @Description Cria um novo tenant com plano e owner. URL code gerado automaticamente. Requer permissão manage_tenants.
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body tenantModels.CreateTenantRequest true "Dados do tenant"
// @Success 201 {object} swagger.CreateTenantResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 409 {object} swagger.ErrorResponse
// @Router /tenants [post]
func (h *Handler) CreateTenant(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	var req tenantModels.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
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
	urlCode := utils.GenerateURLCode()
	tenantID, err := h.service.Repo().CreateTenant(ctx, tx, req.Name, urlCode, req.Subdomain, req.IsCompany, req.CompanyName)
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
		ownerInfo, err = createOwnerForTenant(ctx, tx, tenantID, req.OwnerEmail, req.OwnerFullName, req.OwnerPassword, urlCode)
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
		"tenant": gin.H{"id": tenantID, "name": req.Name, "url_code": urlCode, "status": "active"},
		"owner":  ownerInfo,
	})
}

// GetTenant godoc
// @Summary Obter tenant por ID
// @Description Retorna detalhes de um tenant. Requer permissão view_tenants.
// @Tags Tenants
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do tenant"
// @Success 200 {object} swagger.TenantResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /tenants/{id} [get]
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

// UpdateTenant godoc
// @Summary Atualizar tenant
// @Description Atualiza dados de um tenant. Requer permissão manage_tenants.
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do tenant"
// @Param request body tenantModels.UpdateTenantRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /tenants/{id} [put]
func (h *Handler) UpdateTenant(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	var req tenantModels.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpdateTenant(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "tenant_updated"})
}

// DeleteTenant godoc
// @Summary Remover tenant
// @Description Remove (soft delete) um tenant. Requer permissão manage_tenants.
// @Tags Tenants
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do tenant"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /tenants/{id} [delete]
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

// UpdateTenantStatus godoc
// @Summary Atualizar status do tenant
// @Description Atualiza o status de um tenant (active, suspended, etc). Requer permissão manage_tenants.
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do tenant"
// @Param request body tenantModels.UpdateTenantStatusRequest true "Novo status"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /tenants/{id}/status [put]
func (h *Handler) UpdateTenantStatus(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_tenants") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	id := c.Param("id")
	var req tenantModels.UpdateTenantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpdateTenantStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_status"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "status_updated"})
}

// ChangeTenantPlan godoc
// @Summary Alterar plano do tenant
// @Description Altera o plano de um tenant, com opção de promoção. Requer permissão manage_plans.
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do tenant"
// @Param request body tenantModels.ChangePlanRequest true "Dados do novo plano"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /tenants/{id}/plan [put]
func (h *Handler) ChangeTenantPlan(c *gin.Context) {
	adminID := c.GetString("admin_id")
	if !h.service.HasPermission(c.Request.Context(), adminID, "manage_plans") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission_denied"})
		return
	}

	tenantID := c.Param("id")
	var req tenantModels.ChangePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
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

// GetTenantPlanHistory godoc
// @Summary Histórico de planos do tenant
// @Description Retorna o histórico de planos de um tenant. Requer permissão view_tenants.
// @Tags Tenants
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do tenant"
// @Success 200 {object} swagger.TenantPlanHistoryResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /tenants/{id}/plan-history [get]
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

// GetTenantMembers godoc
// @Summary Listar membros do tenant
// @Description Retorna os membros de um tenant. Requer permissão view_tenants.
// @Tags Tenants
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do tenant"
// @Success 200 {object} swagger.TenantMemberListResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /tenants/{id}/members [get]
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

// ListPlans godoc
// @Summary Listar planos
// @Description Retorna todos os planos com suas features
// @Tags Plans
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.PlanListResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /plans [get]
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.service.Repo().ListPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_plans"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": plans})
}

// CreatePlan godoc
// @Summary Criar plano
// @Description Cria um novo plano com features opcionais
// @Tags Plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body tenantModels.CreatePlanRequest true "Dados do plano"
// @Success 201 {object} swagger.PlanResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /plans [post]
func (h *Handler) CreatePlan(c *gin.Context) {
	var req tenantModels.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	maxUsers := req.MaxUsers
	if maxUsers == 0 {
		maxUsers = 1
	}

	id, err := h.service.Repo().CreatePlan(c.Request.Context(), req.Name, req.Description, req.PlanType, req.Price, maxUsers, req.IsMultilang)
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

// GetPlan godoc
// @Summary Obter plano por ID
// @Description Retorna um plano específico com suas features
// @Tags Plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do plano"
// @Success 200 {object} swagger.PlanResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /plans/{id} [get]
func (h *Handler) GetPlan(c *gin.Context) {
	id := c.Param("id")
	plan, err := h.service.Repo().GetPlanByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan_not_found"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

// UpdatePlan godoc
// @Summary Atualizar plano
// @Description Atualiza um plano existente
// @Tags Plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do plano"
// @Param request body tenantModels.UpdatePlanRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /plans/{id} [put]
func (h *Handler) UpdatePlan(c *gin.Context) {
	id := c.Param("id")
	var req tenantModels.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpdatePlan(c.Request.Context(), id, req.Name, req.Description, req.Price, req.MaxUsers, req.IsMultilang, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_plan"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "plan_updated"})
}

// DeletePlan godoc
// @Summary Remover plano
// @Description Remove um plano do sistema
// @Tags Plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do plano"
// @Success 200 {object} swagger.MessageResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /plans/{id} [delete]
func (h *Handler) DeletePlan(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Repo().DeletePlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete_plan"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "plan_deleted"})
}

// AddFeatureToPlan godoc
// @Summary Adicionar feature ao plano
// @Description Adiciona uma feature a um plano existente
// @Tags Plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do plano"
// @Param request body tenantModels.PlanFeatureRequest true "ID da feature"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /plans/{id}/features [post]
func (h *Handler) AddFeatureToPlan(c *gin.Context) {
	planID := c.Param("id")
	var req tenantModels.PlanFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}
	if err := h.service.Repo().AddFeatureToPlan(c.Request.Context(), planID, req.FeatureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_add_feature"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "feature_added"})
}

// RemoveFeatureFromPlan godoc
// @Summary Remover feature do plano
// @Description Remove uma feature de um plano
// @Tags Plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do plano"
// @Param feat_id path string true "ID da feature"
// @Success 200 {object} swagger.MessageResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /plans/{id}/features/{feat_id} [delete]
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

// ListFeatures godoc
// @Summary Listar features
// @Description Retorna todas as features do sistema
// @Tags Features
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.FeatureListResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /features [get]
func (h *Handler) ListFeatures(c *gin.Context) {
	features, err := h.service.Repo().ListFeatures(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_features"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": features})
}

// CreateFeature godoc
// @Summary Criar feature
// @Description Cria uma nova feature no sistema
// @Tags Features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body tenantModels.CreateFeatureRequest true "Dados da feature"
// @Success 201 {object} swagger.FeatureResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 409 {object} swagger.ErrorResponse
// @Router /features [post]
func (h *Handler) CreateFeature(c *gin.Context) {
	var req tenantModels.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
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

// GetFeature godoc
// @Summary Obter feature por ID
// @Description Retorna uma feature específica
// @Tags Features
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID da feature"
// @Success 200 {object} swagger.FeatureResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /features/{id} [get]
func (h *Handler) GetFeature(c *gin.Context) {
	id := c.Param("id")
	feature, err := h.service.Repo().GetFeatureByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "feature_not_found"})
		return
	}
	c.JSON(http.StatusOK, feature)
}

// UpdateFeature godoc
// @Summary Atualizar feature
// @Description Atualiza uma feature existente
// @Tags Features
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID da feature"
// @Param request body tenantModels.UpdateFeatureRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /features/{id} [put]
func (h *Handler) UpdateFeature(c *gin.Context) {
	id := c.Param("id")
	var req tenantModels.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpdateFeature(c.Request.Context(), id, req.Title, req.Description, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_feature"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "feature_updated"})
}

// DeleteFeature godoc
// @Summary Remover feature
// @Description Remove uma feature do sistema
// @Tags Features
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID da feature"
// @Success 200 {object} swagger.MessageResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /features/{id} [delete]
func (h *Handler) DeleteFeature(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Repo().DeleteFeature(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_delete_feature"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "feature_deleted"})
}

// --- Promotions ---

// ListPromotions godoc
// @Summary Listar promoções
// @Description Retorna todas as promoções do sistema
// @Tags Promotions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.PromotionListResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /promotions [get]
func (h *Handler) ListPromotions(c *gin.Context) {
	promos, err := h.service.Repo().ListPromotions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_list_promotions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": promos})
}

// CreatePromotion godoc
// @Summary Criar promoção
// @Description Cria uma nova promoção
// @Tags Promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body tenantModels.CreatePromotionRequest true "Dados da promoção"
// @Success 201 {object} swagger.PromotionResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /promotions [post]
func (h *Handler) CreatePromotion(c *gin.Context) {
	var req tenantModels.CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
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

// GetPromotion godoc
// @Summary Obter promoção por ID
// @Description Retorna uma promoção específica
// @Tags Promotions
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID da promoção"
// @Success 200 {object} swagger.PromotionResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /promotions/{id} [get]
func (h *Handler) GetPromotion(c *gin.Context) {
	id := c.Param("id")
	promo, err := h.service.Repo().GetPromotionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "promotion_not_found"})
		return
	}
	c.JSON(http.StatusOK, promo)
}

// UpdatePromotion godoc
// @Summary Atualizar promoção
// @Description Atualiza uma promoção existente
// @Tags Promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID da promoção"
// @Param request body tenantModels.UpdatePromotionRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /promotions/{id} [put]
func (h *Handler) UpdatePromotion(c *gin.Context) {
	id := c.Param("id")
	var req tenantModels.UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.service.Repo().UpdatePromotion(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_update_promotion"})
		return
	}
	c.JSON(http.StatusOK, shared.MessageResponse{Message: "promotion_updated"})
}

// DeletePromotion godoc
// @Summary Desativar promoção
// @Description Desativa uma promoção (soft delete)
// @Tags Promotions
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID da promoção"
// @Success 200 {object} swagger.MessageResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /promotions/{id} [delete]
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
