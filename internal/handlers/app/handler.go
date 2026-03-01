package app

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/saas-single-db-api/internal/cache"
	"github.com/saas-single-db-api/internal/i18n"
	_ "github.com/saas-single-db-api/internal/models/swagger"
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

// Register godoc
// @Summary Registrar app user
// @Description Registra um novo usuário no app do tenant
// @Tags Auth
// @Accept json
// @Produce json
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.AppRegisterRequest true "Dados de registro"
// @Success 201 {object} swagger.AppAuthResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /{url_code}/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	urlCode := c.Param("url_code")

	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err, c)})
		return
	}

	result, err := h.service.Register(c.Request.Context(), tenantID, urlCode, req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, err.Error())})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id": result.UserID,
		"token":   result.Token,
	})
}

// Login godoc
// @Summary Login de app user
// @Description Autentica um app user e retorna token JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.AppLoginRequest true "Credenciais"
// @Success 200 {object} swagger.AppAuthResponse
// @Failure 401 {object} swagger.ErrorResponse
// @Router /{url_code}/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	urlCode := c.Param("url_code")

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err, c)})
		return
	}

	result, err := h.service.Login(c.Request.Context(), tenantID, urlCode, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": i18n.T(c, err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   result.Token,
		"user_id": result.UserID,
		"name":    result.Name,
		"email":   result.Email,
	})
}

// Logout godoc
// @Summary Logout de app user
// @Description Invalida o token JWT do app user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {object} swagger.MessageResponse
// @Router /{url_code}/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token != "" {
		h.cache.SetBlacklist(c.Request.Context(), token)
	}
	c.JSON(http.StatusOK, gin.H{"message": i18n.T(c, "logged_out")})
}

// Me godoc
// @Summary Dados do app user logado
// @Description Retorna os dados do app user autenticado
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {object} swagger.AppUserDTO
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("app_user_id")

	result, err := h.service.GetMe(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(c, err.Error())})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ChangePassword godoc
// @Summary Alterar senha do app user
// @Description Altera a senha do app user autenticado
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.AppChangePasswordRequest true "Senhas"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /{url_code}/auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("app_user_id")

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err, c)})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), tenantID, userID, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": i18n.T(c, "password_changed")})
}

// ForgotPassword godoc
// @Summary Solicitar reset de senha
// @Description Envia email com link para reset de senha
// @Tags Auth
// @Accept json
// @Produce json
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.AppForgotPasswordRequest true "Email"
// @Success 200 {object} swagger.MessageResponse
// @Router /{url_code}/auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err, c)})
		return
	}

	_ = h.service.ForgotPassword(c.Request.Context(), tenantID, req.Email)
	c.JSON(http.StatusOK, gin.H{"message": i18n.T(c, "password_reset_sent")})
}

// ResetPassword godoc
// @Summary Resetar senha
// @Description Reseta a senha usando token recebido por email
// @Tags Auth
// @Accept json
// @Produce json
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.AppResetPasswordRequest true "Token e nova senha"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /{url_code}/auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err, c)})
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), tenantID, req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": i18n.T(c, "password_reset_success")})
}

// ==================== PROFILE ====================

// GetProfile godoc
// @Summary Obter perfil do app user
// @Description Retorna o perfil completo do app user autenticado
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Success 200 {object} swagger.AppUserProfileDTO
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("app_user_id")
	profile, err := h.repo.GetAppUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(c, "profile_not_found")})
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

// UpdateProfile godoc
// @Summary Atualizar perfil do app user
// @Description Atualiza campos do perfil do app user
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param request body swagger.AppUpdateProfileRequest true "Dados do perfil"
// @Success 200 {object} swagger.MessageResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /{url_code}/profile [put]
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
		c.JSON(http.StatusBadRequest, gin.H{"error": utils.FormatValidationErrors(err, c)})
		return
	}

	if err := h.repo.UpdateAppUserProfile(c.Request.Context(), userID, req.FullName, req.Phone, req.Document, req.Address, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "failed_update_profile")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": i18n.T(c, "profile_updated")})
}

// UploadAvatar godoc
// @Summary Upload de avatar do app user
// @Description Faz upload da foto de perfil do app user
// @Tags Profile
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param url_code path string true "URL code do tenant"
// @Param avatar formData file true "Imagem do avatar"
// @Success 200 {object} swagger.UploadResponse
// @Failure 400 {object} swagger.ErrorResponse
// @Router /{url_code}/profile/avatar [put]
func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := c.GetString("app_user_id")
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, "no_file_provided")})
		return
	}
	defer file.Close()

	publicURL, storagePath, err := h.storage.Upload(file, header, "app-avatars")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "failed_upload")})
		return
	}

	_ = h.repo.UpdateAppUserProfileAvatar(c.Request.Context(), userID, publicURL)
	c.JSON(http.StatusOK, gin.H{"path": storagePath, "public_url": publicURL})
}

// ==================== CATALOG (Public) ====================

// ListProducts godoc
// @Summary Listar produtos
// @Description Retorna produtos ativos do tenant paginados
// @Tags Catalog
// @Produce json
// @Param url_code path string true "URL code do tenant"
// @Param page query int false "Página" default(1)
// @Param page_size query int false "Itens por página" default(20)
// @Success 200 {object} swagger.PaginatedResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/catalog/products [get]
func (h *Handler) ListProducts(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	pag := utils.GetPagination(c)

	products, total, err := h.repo.ListActiveProducts(c.Request.Context(), tenantID, pag.PageSize, pag.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "failed_list_products")})
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
// @Description Retorna detalhes de um produto ativo
// @Tags Catalog
// @Produce json
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do produto"
// @Success 200 {object} swagger.ProductResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/catalog/products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	productID := c.Param("id")
	product, err := h.repo.GetActiveProduct(c.Request.Context(), tenantID, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(c, "product_not_found")})
		return
	}
	c.JSON(http.StatusOK, product)
}

// ListServices godoc
// @Summary Listar serviços
// @Description Retorna serviços ativos do tenant paginados
// @Tags Catalog
// @Produce json
// @Param url_code path string true "URL code do tenant"
// @Param page query int false "Página" default(1)
// @Param page_size query int false "Itens por página" default(20)
// @Success 200 {object} swagger.PaginatedResponse
// @Failure 500 {object} swagger.ErrorResponse
// @Router /{url_code}/catalog/services [get]
func (h *Handler) ListServices(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	pag := utils.GetPagination(c)

	services, total, err := h.repo.ListActiveServices(c.Request.Context(), tenantID, pag.PageSize, pag.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "failed_list_services")})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      services,
		"total":     total,
		"page":      pag.Page,
		"page_size": pag.PageSize,
	})
}

// GetServiceDetail godoc
// @Summary Obter serviço
// @Description Retorna detalhes de um serviço ativo
// @Tags Catalog
// @Produce json
// @Param url_code path string true "URL code do tenant"
// @Param id path string true "ID do serviço"
// @Success 200 {object} swagger.ServiceResponse
// @Failure 404 {object} swagger.ErrorResponse
// @Router /{url_code}/catalog/services/{id} [get]
func (h *Handler) GetServiceDetail(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	serviceID := c.Param("id")
	service, err := h.repo.GetActiveService(c.Request.Context(), tenantID, serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.T(c, "service_not_found")})
		return
	}
	c.JSON(http.StatusOK, service)
}
