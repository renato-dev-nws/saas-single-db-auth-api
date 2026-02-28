package tenant

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/saas-single-db-api/internal/cache"
	_ "github.com/saas-single-db-api/internal/models/swagger"
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

// ListPlans godoc
// @Summary Listar planos ativos
// @Description Retorna todos os planos ativos disponíveis para assinatura
// @Tags Plans
// @Produce json
// @Success 200 {array} swagger.PlanResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /plans [get]
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.repo.ListActivePlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list plans"})
		return
	}
	c.JSON(http.StatusOK, plans)
}

// Subscribe godoc
// @Summary Criar assinatura (self-service)
// @Description Cria um novo tenant com plano, owner e gera url_code automaticamente. Envia email de verificação.
// @Tags Subscription
// @Accept json
// @Produce json
// @Param request body swagger.SubscribeRequest true "Dados da assinatura"
// @Success 201 {object} swagger.SubscriptionResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /subscription [post]
func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		Name         string  `json:"name" binding:"required"`
		Email        string  `json:"email" binding:"required,email"`
		Password     string  `json:"password" binding:"required,min=6"`
		IsCompany    bool    `json:"is_company"`
		CompanyName  string  `json:"company_name"`
		PlanID       string  `json:"plan_id" binding:"required"`
		BillingCycle string  `json:"billing_cycle" binding:"required"`
		PromoCode    *string `json:"promo_code"`
		Subdomain    string  `json:"subdomain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	// company_name is required when is_company is true
	if req.IsCompany && strings.TrimSpace(req.CompanyName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": map[string]string{
			"company_name": "Company name is required when is_company is true",
		}})
		return
	}

	// Determine tenant name: company_name if company, otherwise user name
	tenantName := req.Name
	if req.IsCompany && req.CompanyName != "" {
		tenantName = req.CompanyName
	}

	subdomain := strings.ToLower(strings.ReplaceAll(req.Subdomain, " ", ""))

	result, err := h.service.Subscribe(c.Request.Context(), svc.SubscribeInput{
		TenantName:   tenantName,
		Subdomain:    subdomain,
		IsCompany:    req.IsCompany,
		CompanyName:  req.CompanyName,
		PlanID:       req.PlanID,
		BillingCycle: req.BillingCycle,
		PromoCode:    req.PromoCode,
		OwnerName:    req.Name,
		OwnerEmail:   req.Email,
		OwnerPass:    req.Password,
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

// VerifyEmail godoc
// @Summary Verificar email
// @Description Valida o token de verificação de email e marca o email como verificado
// @Tags Auth
// @Produce json
// @Param token query string true "Token de verificação"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /auth/verify-email [get]
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

// ResendVerification godoc
// @Summary Reenviar email de verificação
// @Description Reenvia o email de verificação para o usuário autenticado
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /auth/resend-verification [post]
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

// Login godoc
// @Summary Login de usuário
// @Description Autentica um usuário do backoffice com email e senha
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body swagger.UserLoginRequest true "Credenciais"
// @Success 200 {object} swagger.UserLoginResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":               result.Token,
		"name":                result.Name,
		"email":               result.Email,
		"current_tenant_code": result.CurrentTenantCode,
		"tenants":             result.Tenants,
	})
}

// Logout godoc
// @Summary Logout de usuário
// @Description Invalida o token JWT do usuário
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.MessageResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token != "" {
		h.cache.SetBlacklist(c.Request.Context(), token)
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// Me godoc
// @Summary Dados do usuário autenticado
// @Description Retorna os dados do usuário logado com seus tenants
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.BackofficeUserDTO
// @Failure 404 {object} swagger.ErrorResponse
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	result, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ChangePassword godoc
// @Summary Alterar senha
// @Description Altera a senha do usuário autenticado
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body swagger.ChangePasswordRequest true "Senhas atual e nova"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /profile/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
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

// SelectTenant godoc
// @Summary Selecionar tenant
// @Description Troca o contexto do usuário para outro tenant, retornando um novo token com escopo
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {object} swagger.SwitchTenantResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /auth/switch/{url_code} [post]
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

// GetProfile godoc
// @Summary Obter perfil do usuário
// @Description Retorna o perfil do usuário autenticado
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} swagger.UserProfileDTO
// @Failure 404 {object} swagger.ErrorResponse
// @Router /profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	profile, err := h.repo.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary Atualizar perfil do usuário
// @Description Atualiza o perfil do usuário autenticado
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body swagger.UpdateUserProfileRequest true "Dados do perfil"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		FullName *string `json:"full_name"`
		About    *string `json:"about"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}
	if err := h.repo.UpdateUserProfile(c.Request.Context(), userID, req.FullName, req.About); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

// UploadAvatar godoc
// @Summary Upload de avatar
// @Description Faz upload do avatar do usuário
// @Tags Profile
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Imagem do avatar"
// @Success 200 {object} swagger.UploadResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /profile/avatar [post]
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

// ==================== BOOTSTRAP ====================

// GetBootstrap godoc
// @Summary Obter bootstrap do tenant
// @Description Retorna dados essenciais para inicializar o frontend: tenant, features, permissões, plano e layout
// @Tags Bootstrap
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {object} swagger.BootstrapResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/bootstrap [get]
func (h *Handler) GetBootstrap(c *gin.Context) {
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
	isOwner := h.service.IsOwner(c.Request.Context(), userID, tenantID)

	// Layout settings — return defaults if none saved yet
	layoutData := map[string]interface{}{
		"primary_color":   "#4F46E5",
		"secondary_color": "#10B981",
		"logo":            "",
		"theme":           "Aura",
	}
	layoutSetting, err := h.repo.GetSetting(c.Request.Context(), tenantID, "layout")
	if err == nil && layoutSetting != nil {
		if m, ok := layoutSetting.Data.(map[string]interface{}); ok {
			layoutData = m
		}
	}

	result := gin.H{
		"tenant":          tenant,
		"features":        features,
		"permissions":     perms,
		"is_owner":        isOwner,
		"layout_settings": layoutData,
	}

	if plan != nil {
		result["plan"] = gin.H{
			"name":         plan.PlanName,
			"max_users":    plan.MaxUsers,
			"is_multilang": plan.IsMultilang,
		}
	}

	c.JSON(http.StatusOK, result)
}

// ==================== TENANT PROFILE ====================

// GetTenantProfile godoc
// @Summary Obter perfil do tenant
// @Description Retorna o perfil do tenant (about, logo, custom_settings)
// @Tags Tenant Profile
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {object} swagger.TenantProfileResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/tenant [get]
func (h *Handler) GetTenantProfile(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	profile, err := h.repo.GetTenantProfile(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// UpdateTenantProfile godoc
// @Summary Atualizar perfil do tenant
// @Description Atualiza o perfil do tenant. Apenas o owner pode atualizar.
// @Tags Tenant Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.UpdateTenantProfileRequest true "Dados do perfil"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/tenant/profile [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.repo.UpdateTenantProfile(c.Request.Context(), tenantID, req.About, req.CustomSettings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tenant profile updated"})
}

// UploadLogo godoc
// @Summary Upload de logo do tenant
// @Description Faz upload do logo do tenant. Apenas o owner pode enviar.
// @Tags Tenant Profile
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param logo formData file true "Imagem do logo"
// @Success 200 {object} swagger.UploadResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/tenant/logo [post]
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

// ListMembers godoc
// @Summary Listar membros do tenant
// @Description Retorna todos os membros do tenant
// @Tags Members
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {array} swagger.MemberDTO
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	members, err := h.repo.ListTenantMembers(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list members"})
		return
	}
	c.JSON(http.StatusOK, members)
}

// GetMember godoc
// @Summary Obter membro por ID
// @Description Retorna detalhes de um membro específico
// @Tags Members
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do membro"
// @Success 200 {object} swagger.MemberDTO
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/members/{id} [get]
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

// CanAddMember godoc
// @Summary Verificar se pode adicionar membro
// @Description Verifica se o tenant pode adicionar mais membros baseado no plano
// @Tags Members
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {object} swagger.CanAddMemberResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/members/can-add [get]
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

// InviteMember godoc
// @Summary Convidar membro
// @Description Adiciona um novo membro ao tenant. Requer permissão user_m ou ser owner.
// @Tags Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.InviteMemberRequest true "Dados do membro"
// @Success 201 {object} swagger.InviteMemberResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/members [post]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	newUserID, err := h.service.InviteMember(c.Request.Context(), tenantID, req.Email, req.Name, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": newUserID, "message": "member added"})
}

// UpdateMemberRole godoc
// @Summary Atualizar role do membro
// @Description Atualiza a role de um membro. Requer permissão user_m ou ser owner.
// @Tags Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do membro"
// @Param request body swagger.UpdateMemberRoleRequest true "Nova role"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/members/{id}/role [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.repo.UpdateMemberRole(c.Request.Context(), tenantID, memberID, req.RoleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member role updated"})
}

// RemoveMember godoc
// @Summary Remover membro
// @Description Remove um membro do tenant. Apenas o owner pode remover.
// @Tags Members
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do membro"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/members/{id} [delete]
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

// ListRoles godoc
// @Summary Listar roles do tenant
// @Description Retorna todas as roles do tenant
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {array} swagger.UserRoleResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	roles, err := h.repo.ListTenantRoles(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// GetRole godoc
// @Summary Obter role com permissões
// @Description Retorna uma role do tenant com suas permissões
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID da role"
// @Success 200 {object} swagger.RoleDetailResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/roles/{id} [get]
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

// CreateRole godoc
// @Summary Criar role
// @Description Cria uma nova role no tenant. Apenas o owner pode criar.
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.CreateRoleRequest true "Dados da role"
// @Success 201 {object} swagger.UserRoleResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/roles [post]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
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

// UpdateRole godoc
// @Summary Atualizar role
// @Description Atualiza uma role do tenant. Apenas o owner pode atualizar.
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID da role"
// @Param request body swagger.UpdateRoleRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/roles/{id} [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.repo.UpdateTenantRole(c.Request.Context(), tenantID, roleID, req.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

// DeleteRole godoc
// @Summary Remover role
// @Description Remove uma role do tenant. Apenas o owner pode remover.
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID da role"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/roles/{id} [delete]
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

// AssignPermission godoc
// @Summary Atribuir permissão a role
// @Description Atribui uma permissão a uma role do tenant. Apenas o owner.
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID da role"
// @Param request body swagger.AssignPermissionRequest true "ID da permissão"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/roles/{id}/permissions [post]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.repo.AssignPermissionToRole(c.Request.Context(), roleID, req.PermissionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign permission"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permission assigned"})
}

// RemovePermission godoc
// @Summary Remover permissão da role
// @Description Remove uma permissão de uma role do tenant. Apenas o owner.
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID da role"
// @Param permId path string true "ID da permissão"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/roles/{id}/permissions/{permId} [delete]
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

// ListProducts godoc
// @Summary Listar produtos
// @Description Retorna produtos do tenant paginados. Requer feature 'products'.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param page query int false "Página" default(1)
// @Param page_size query int false "Itens por página" default(20)
// @Success 200 {object} swagger.PaginatedResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/products [get]
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

// GetProduct godoc
// @Summary Obter produto
// @Description Retorna um produto específico. Requer feature 'products'.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do produto"
// @Success 200 {object} swagger.ProductResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/products/{id} [get]
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

// CreateProduct godoc
// @Summary Criar produto
// @Description Cria um novo produto. Requer feature 'products' e permissão 'prod_c'.
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.CreateProductRequest true "Dados do produto"
// @Success 201 {object} swagger.CreateIDResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/products [post]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	id, err := h.repo.CreateProduct(c.Request.Context(), tenantID, req.Name, req.Description, req.Price, req.SKU, req.Stock)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// UpdateProduct godoc
// @Summary Atualizar produto
// @Description Atualiza um produto. Requer feature 'products' e permissão 'prod_u'.
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do produto"
// @Param request body swagger.UpdateProductRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/products/{id} [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.repo.UpdateProduct(c.Request.Context(), tenantID, productID, req.Name, req.Description, req.Price, req.SKU, req.Stock, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

// DeleteProduct godoc
// @Summary Remover produto
// @Description Remove um produto. Requer feature 'products' e permissão 'prod_d'.
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do produto"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/products/{id} [delete]
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

// UploadProductImage godoc
// @Summary Upload de imagem do produto
// @Description Faz upload de imagem para um produto. Requer feature 'products' e permissão 'prod_u'.
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do produto"
// @Param image formData file true "Imagem do produto"
// @Success 200 {object} swagger.UploadResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/products/{id}/image [post]
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

// ListServices godoc
// @Summary Listar serviços
// @Description Retorna serviços do tenant paginados. Requer feature 'services'.
// @Tags Services
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param page query int false "Página" default(1)
// @Param page_size query int false "Itens por página" default(20)
// @Success 200 {object} swagger.PaginatedResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/services [get]
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

// GetService godoc
// @Summary Obter serviço
// @Description Retorna um serviço específico. Requer feature 'services'.
// @Tags Services
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do serviço"
// @Success 200 {object} swagger.ServiceResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/services/{id} [get]
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

// CreateService godoc
// @Summary Criar serviço
// @Description Cria um novo serviço. Requer feature 'services' e permissão 'serv_c'.
// @Tags Services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.CreateServiceRequest true "Dados do serviço"
// @Success 201 {object} swagger.CreateIDResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/services [post]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	id, err := h.repo.CreateService(c.Request.Context(), tenantID, req.Name, req.Description, req.Price, req.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create service"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// UpdateService godoc
// @Summary Atualizar serviço
// @Description Atualiza um serviço. Requer feature 'services' e permissão 'serv_u'.
// @Tags Services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do serviço"
// @Param request body swagger.UpdateServiceRequest true "Dados para atualização"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/services/{id} [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.repo.UpdateService(c.Request.Context(), tenantID, serviceID, req.Name, req.Description, req.Price, req.Duration, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "service updated"})
}

// DeleteService godoc
// @Summary Remover serviço
// @Description Remove um serviço. Requer feature 'services' e permissão 'serv_d'.
// @Tags Services
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do serviço"
// @Success 200 {object} swagger.MessageResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/services/{id} [delete]
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

// UploadServiceImage godoc
// @Summary Upload de imagem do serviço
// @Description Faz upload de imagem para um serviço. Requer feature 'services' e permissão 'serv_u'.
// @Tags Services
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do serviço"
// @Param image formData file true "Imagem do serviço"
// @Success 200 {object} swagger.UploadResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/services/{id}/image [post]
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

// GetLayoutSettings godoc
// @Summary Obter configurações de layout
// @Description Retorna as configurações de layout do tenant (cores, logo, tema)
// @Tags Settings
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {object} swagger.LayoutSettingsResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/settings/layout [get]
func (h *Handler) GetLayoutSettings(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	// Default layout settings
	layoutData := map[string]interface{}{
		"primary_color":   "#4F46E5",
		"secondary_color": "#10B981",
		"logo":            "",
		"theme":           "Aura",
	}

	setting, err := h.repo.GetSetting(c.Request.Context(), tenantID, "layout")
	if err == nil && setting != nil {
		if m, ok := setting.Data.(map[string]interface{}); ok {
			layoutData = m
		}
	}

	c.JSON(http.StatusOK, layoutData)
}

// UpdateLayoutSettings godoc
// @Summary Atualizar configurações de layout
// @Description Atualiza cores, logo e tema do tenant. Requer permissão 'setg_m' ou ser owner.
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.LayoutSettingsRequest true "Configurações de layout"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/settings/layout [put]
func (h *Handler) UpdateLayoutSettings(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")

	if !h.service.HasPermission(c.Request.Context(), userID, tenantID, "setg_m") &&
		!h.service.IsOwner(c.Request.Context(), userID, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	var req struct {
		PrimaryColor   string `json:"primary_color" binding:"required"`
		SecondaryColor string `json:"secondary_color" binding:"required"`
		Logo           string `json:"logo"`
		Theme          string `json:"theme" binding:"required,oneof=Aura Lara"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	data := map[string]interface{}{
		"primary_color":   req.PrimaryColor,
		"secondary_color": req.SecondaryColor,
		"logo":            req.Logo,
		"theme":           req.Theme,
	}

	if err := h.repo.UpsertSetting(c.Request.Context(), tenantID, "layout", data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save layout settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "layout_settings_saved"})
}

// ListSettings godoc
// @Summary Listar configurações
// @Description Retorna todas as configurações do tenant
// @Tags Settings
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {array} swagger.SettingResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/settings [get]
func (h *Handler) ListSettings(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	settings, err := h.repo.ListSettings(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list settings"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// GetSetting godoc
// @Summary Obter configuração por categoria
// @Description Retorna uma configuração específica do tenant por categoria
// @Tags Settings
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param category path string true "Categoria da configuração"
// @Success 200 {object} swagger.SettingResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/settings/{category} [get]
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

// UpsertSetting godoc
// @Summary Salvar configuração
// @Description Cria ou atualiza uma configuração. Requer permissão 'setg_m' ou ser owner.
// @Tags Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param category path string true "Categoria da configuração"
// @Param request body swagger.UpsertSettingRequest true "Dados da configuração"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Failure 403 {object} swagger.ErrorResponse
// @Router /{url_code}/settings/{category} [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}

	if err := h.repo.UpsertSetting(c.Request.Context(), tenantID, req.Category, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "setting saved"})
}

// ==================== IMAGES ====================

// ListImages godoc
// @Summary Listar imagens
// @Description Retorna imagens do tenant paginadas
// @Tags Images
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param page query int false "Página" default(1)
// @Param page_size query int false "Itens por página" default(20)
// @Success 200 {object} swagger.PaginatedResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/images [get]
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

// UploadImage godoc
// @Summary Upload de imagem genérica
// @Description Faz upload de uma imagem com entity_type e entity_id opcionais
// @Tags Images
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param image formData file true "Imagem"
// @Param entity_type formData string false "Tipo da entidade"
// @Param entity_id formData string false "ID da entidade"
// @Success 201 {object} swagger.ImageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /{url_code}/images [post]
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

// DeleteImage godoc
// @Summary Remover imagem
// @Description Remove uma imagem do tenant
// @Tags Images
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID da imagem"
// @Success 200 {object} swagger.MessageResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/images/{id} [delete]
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

// ListAppUsers godoc
// @Summary Listar app users
// @Description Retorna app users do tenant paginados
// @Tags App Users
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param page query int false "Página" default(1)
// @Param page_size query int false "Itens por página" default(20)
// @Success 200 {object} swagger.PaginatedResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/app-users [get]
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

// GetAppUser godoc
// @Summary Obter app user
// @Description Retorna um app user específico
// @Tags App Users
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do app user"
// @Success 200 {object} swagger.AppUserListDTO
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/app-users/{id} [get]
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

// UpdateAppUserStatus godoc
// @Summary Atualizar status do app user
// @Description Atualiza o status de um app user (active/blocked)
// @Tags App Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do app user"
// @Param request body swagger.UpdateAppUserStatusRequest true "Novo status"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /{url_code}/app-users/{id}/status [put]
func (h *Handler) UpdateAppUserStatus(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	appUserID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required,oneof=active blocked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err)})
		return
	}
	if err := h.repo.UpdateAppUserStatus(c.Request.Context(), tenantID, appUserID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// DeleteAppUser godoc
// @Summary Remover app user
// @Description Remove (soft delete) um app user
// @Tags App Users
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do app user"
// @Success 200 {object} swagger.MessageResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/app-users/{id} [delete]
func (h *Handler) DeleteAppUser(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	appUserID := c.Param("id")
	if err := h.repo.SoftDeleteAppUser(c.Request.Context(), tenantID, appUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete app user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "app user deleted"})
}
