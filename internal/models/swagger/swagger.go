// Package swagger contains supplementary models used exclusively for Swagger documentation.
// These types describe response shapes that handlers build via gin.H{} maps.
package swagger

import (
	"time"
)

// ─── Generic ─────────────────────────────────────────────

// TokenResponse represents a JWT token response
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// ErrorResponse represents an error response (string for general errors)
type ErrorResponse struct {
	Error string `json:"error" example:"error_message"`
}

// ValidationErrorResponse represents a validation error response with field-level errors
type ValidationErrorResponse struct {
	Error map[string]string `json:"error"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"operation_successful"`
}

// DataListResponse represents a generic data list wrapper
type DataListResponse struct {
	Data interface{} `json:"data"`
}

// PaginatedResponse represents a paginated list response
type PaginatedResponse struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total" example:"100"`
	Page     int         `json:"page" example:"1"`
	PageSize int         `json:"page_size" example:"20"`
}

// ─── Admin API ───────────────────────────────────────────

// AdminLoginResponse is the response from admin login
type AdminLoginResponse struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	Admin AdminUserDTO `json:"admin"`
}

// AdminUserDTO is a safe representation of an admin user
type AdminUserDTO struct {
	ID      string           `json:"id" example:"uuid"`
	Email   string           `json:"email" example:"admin@example.com"`
	Name    *string          `json:"name" example:"John Admin"`
	Status  string           `json:"status" example:"active"`
	Profile *AdminProfileDTO `json:"profile,omitempty"`
	Roles   []AdminRoleDTO   `json:"roles,omitempty"`
}

// AdminProfileDTO is the admin profile
type AdminProfileDTO struct {
	FullName    *string     `json:"full_name" example:"John Admin"`
	Title       *string     `json:"title" example:"System Administrator"`
	Bio         *string     `json:"bio" example:"Bio text"`
	AvatarURL   *string     `json:"avatar_url" example:"https://example.com/avatar.jpg"`
	SocialLinks interface{} `json:"social_links"`
}

// AdminRoleDTO is the admin role
type AdminRoleDTO struct {
	ID          int     `json:"id" example:"1"`
	Title       string  `json:"title" example:"Super Admin"`
	Slug        string  `json:"slug" example:"super_admin"`
	Description *string `json:"description,omitempty" example:"Full system access"`
}

// AdminPermissionDTO is the admin permission
type AdminPermissionDTO struct {
	ID          int     `json:"id" example:"1"`
	Title       string  `json:"title" example:"Manage Users"`
	Slug        string  `json:"slug" example:"manage_sys_users"`
	Description *string `json:"description" example:"Can manage system users"`
}

// SysUserDetailResponse is the detailed view of a sys user
type SysUserDetailResponse struct {
	ID      string           `json:"id" example:"uuid"`
	Email   string           `json:"email" example:"admin@example.com"`
	Name    *string          `json:"name" example:"John"`
	Status  string           `json:"status" example:"active"`
	Profile *AdminProfileDTO `json:"profile"`
	Roles   []AdminRoleDTO   `json:"roles"`
}

// AdminRoleListResponse wraps a list of roles
type AdminRoleListResponse struct {
	Data []AdminRoleDTO `json:"data"`
}

// AdminPermissionListResponse wraps a list of permissions
type AdminPermissionListResponse struct {
	Data []AdminPermissionDTO `json:"data"`
}

// ─── Tenant/Plan/Feature/Promotion ───────────────────────

// TenantResponse represents a tenant entity
type TenantResponse struct {
	ID           string    `json:"id" example:"uuid"`
	Name         string    `json:"name" example:"My Company"`
	URLCode      string    `json:"url_code" example:"HRP1ZYERFVA"`
	Subdomain    string    `json:"subdomain" example:"mycompany"`
	IsCompany    bool      `json:"is_company" example:"true"`
	CompanyName  *string   `json:"company_name" example:"My Company Ltd"`
	CustomDomain *string   `json:"custom_domain"`
	Status       string    `json:"status" example:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateTenantResponse is the response from admin tenant creation
type CreateTenantResponse struct {
	Tenant TenantResponse `json:"tenant"`
	Owner  OwnerBrief     `json:"owner"`
}

// OwnerBrief is a brief owner representation
type OwnerBrief struct {
	ID    string `json:"id" example:"uuid"`
	Email string `json:"email" example:"owner@example.com"`
}

// PlanResponse represents a plan with features
type PlanResponse struct {
	ID          string            `json:"id" example:"uuid"`
	Name        string            `json:"name" example:"Business Pro"`
	Description *string           `json:"description" example:"Business plan with all features"`
	PlanType    string            `json:"plan_type" example:"business"`
	Price       float64           `json:"price" example:"99.90"`
	MaxUsers    int               `json:"max_users" example:"5"`
	IsMultilang bool              `json:"is_multilang" example:"true"`
	IsActive    bool              `json:"is_active" example:"true"`
	Features    []FeatureResponse `json:"features,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PlanListResponse wraps a list of plans
type PlanListResponse struct {
	Data []PlanResponse `json:"data"`
}

// FeatureResponse represents a feature entity
type FeatureResponse struct {
	ID          string    `json:"id" example:"uuid"`
	Title       string    `json:"title" example:"Products"`
	Slug        string    `json:"slug" example:"products"`
	Code        string    `json:"code" example:"products"`
	Description *string   `json:"description" example:"Product management feature"`
	IsActive    bool      `json:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FeatureListResponse wraps a list of features
type FeatureListResponse struct {
	Data []FeatureResponse `json:"data"`
}

// PromotionResponse represents a promotion entity
type PromotionResponse struct {
	ID             string     `json:"id" example:"uuid"`
	Name           string     `json:"name" example:"Launch Discount"`
	Description    *string    `json:"description" example:"50% off for 3 months"`
	DiscountType   string     `json:"discount_type" example:"percentage"`
	DiscountValue  float64    `json:"discount_value" example:"50"`
	DurationMonths int        `json:"duration_months" example:"3"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until"`
	IsActive       bool       `json:"is_active" example:"true"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PromotionListResponse wraps a list of promotions
type PromotionListResponse struct {
	Data []PromotionResponse `json:"data"`
}

// TenantPlanHistoryResponse wraps plan history
type TenantPlanHistoryResponse struct {
	Data []TenantPlanDTO `json:"data"`
}

// TenantPlanDTO represents a tenant plan subscription
type TenantPlanDTO struct {
	ID              string     `json:"id" example:"uuid"`
	TenantID        string     `json:"tenant_id" example:"uuid"`
	PlanID          string     `json:"plan_id" example:"uuid"`
	BillingCycle    string     `json:"billing_cycle" example:"monthly"`
	BasePrice       float64    `json:"base_price" example:"99.90"`
	ContractedPrice float64    `json:"contracted_price" example:"99.90"`
	PromotionID     *string    `json:"promotion_id"`
	PromoPrice      *float64   `json:"promo_price"`
	PromoExpiresAt  *time.Time `json:"promo_expires_at"`
	IsActive        bool       `json:"is_active" example:"true"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
}

// TenantMemberListResponse wraps tenant members
type TenantMemberListResponse struct {
	Data []TenantMemberDTO `json:"data"`
}

// TenantMemberDTO represents a tenant member
type TenantMemberDTO struct {
	UserID  string `json:"user_id" example:"uuid"`
	Email   string `json:"email" example:"user@example.com"`
	Name    string `json:"name" example:"John"`
	IsOwner bool   `json:"is_owner" example:"false"`
}

// ─── Tenant API (Backoffice) ─────────────────────────────

// SubscriptionResponse is the response for public subscription
type SubscriptionResponse struct {
	Tenant       TenantBriefDTO      `json:"tenant"`
	Subscription SubscriptionInfoDTO `json:"subscription"`
	Token        string              `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User         UserBriefDTO        `json:"user"`
}

// TenantBriefDTO is a brief tenant representation
type TenantBriefDTO struct {
	ID        string `json:"id" example:"uuid"`
	Name      string `json:"name" example:"My Company"`
	URLCode   string `json:"url_code" example:"HRP1ZYERFVA"`
	Subdomain string `json:"subdomain" example:"mycompany"`
	Status    string `json:"status" example:"active"`
}

// SubscriptionInfoDTO contains plan subscription details
type SubscriptionInfoDTO struct {
	Plan            string     `json:"plan" example:"Business Pro"`
	BillingCycle    string     `json:"billing_cycle" example:"monthly"`
	ContractedPrice float64    `json:"contracted_price" example:"99.90"`
	PromoPrice      *float64   `json:"promo_price,omitempty"`
	PromoExpiresAt  *time.Time `json:"promo_expires_at,omitempty"`
	Promotion       string     `json:"promotion,omitempty"`
}

// UserBriefDTO is a brief user representation
type UserBriefDTO struct {
	ID    string `json:"id" example:"uuid"`
	Email string `json:"email" example:"user@example.com"`
}

// UserLoginResponse is the response for backoffice user login
type UserLoginResponse struct {
	Token         string            `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User          BackofficeUserDTO `json:"user"`
	CurrentTenant *TenantInfoDTO    `json:"current_tenant"`
	Tenants       []TenantBrief     `json:"tenants"`
}

// BackofficeUserDTO is a safe representation of a backoffice user
type BackofficeUserDTO struct {
	ID      string          `json:"id" example:"uuid"`
	Email   string          `json:"email" example:"user@example.com"`
	Name    string          `json:"name" example:"John"`
	Status  string          `json:"status" example:"active"`
	Profile *UserProfileDTO `json:"profile,omitempty"`
}

// UserProfileDTO is the backoffice user profile
type UserProfileDTO struct {
	FullName  *string `json:"full_name" example:"John Doe"`
	About     *string `json:"about" example:"About me"`
	AvatarURL *string `json:"avatar_url" example:"https://example.com/avatar.jpg"`
}

// TenantInfoDTO includes features and permissions for a specific tenant
type TenantInfoDTO struct {
	ID          string   `json:"id" example:"uuid"`
	URLCode     string   `json:"url_code" example:"HRP1ZYERFVA"`
	CompanyName *string  `json:"company_name"`
	Name        string   `json:"name" example:"My Company"`
	Features    []string `json:"features"`
	Permissions []string `json:"permissions"`
}

// TenantBrief is a brief tenant info
type TenantBrief struct {
	ID      string `json:"id" example:"uuid"`
	URLCode string `json:"url_code" example:"HRP1ZYERFVA"`
	Name    string `json:"name" example:"My Company"`
	IsOwner bool   `json:"is_owner" example:"true"`
}

// SwitchTenantResponse is the response for switching tenant
type SwitchTenantResponse struct {
	Token    string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TenantID string `json:"tenant_id" example:"uuid"`
	URLCode  string `json:"url_code" example:"HRP1ZYERFVA"`
}

// TenantConfigResponse is the full config response
type TenantConfigResponse struct {
	Tenant      TenantConfigDTO `json:"tenant"`
	Features    []string        `json:"features"`
	Permissions []string        `json:"permissions"`
	Plan        PlanConfigDTO   `json:"plan"`
}

// TenantConfigDTO is the tenant config info
type TenantConfigDTO struct {
	ID          string  `json:"id" example:"uuid"`
	Name        string  `json:"name" example:"My Company"`
	URLCode     string  `json:"url_code" example:"HRP1ZYERFVA"`
	CompanyName *string `json:"company_name"`
}

// PlanConfigDTO is the plan config info
type PlanConfigDTO struct {
	Name            string     `json:"name" example:"Business Pro"`
	MaxUsers        int        `json:"max_users" example:"5"`
	CurrentUsers    int        `json:"current_users" example:"2"`
	AvailableSlots  int        `json:"available_slots" example:"3"`
	IsMultilang     bool       `json:"is_multilang" example:"true"`
	BillingCycle    string     `json:"billing_cycle" example:"monthly"`
	ContractedPrice float64    `json:"contracted_price" example:"99.90"`
	ActivePrice     float64    `json:"active_price" example:"99.90"`
	PromoExpiresAt  *time.Time `json:"promo_expires_at"`
	PriceUpdatedAt  time.Time  `json:"price_updated_at"`
}

// TenantProfileResponse represents a tenant profile
type TenantProfileResponse struct {
	TenantID       string      `json:"tenant_id" example:"uuid"`
	About          *string     `json:"about" example:"About our company"`
	LogoURL        *string     `json:"logo_url" example:"https://example.com/logo.jpg"`
	CustomSettings interface{} `json:"custom_settings"`
}

// CanAddMemberResponse returns whether a member can be added
type CanAddMemberResponse struct {
	CanAdd         bool   `json:"can_add" example:"true"`
	CurrentUsers   int    `json:"current_users" example:"2"`
	MaxUsers       int    `json:"max_users" example:"5"`
	AvailableSlots int    `json:"available_slots" example:"3"`
	Reason         string `json:"reason,omitempty"`
	UpgradeHint    string `json:"upgrade_hint,omitempty"`
}

// MemberDTO represents a member
type MemberDTO struct {
	UserID    string          `json:"user_id" example:"uuid"`
	Email     string          `json:"email" example:"member@example.com"`
	Name      string          `json:"name" example:"Jane Doe"`
	IsOwner   bool            `json:"is_owner" example:"false"`
	Role      *RoleBriefDTO   `json:"role"`
	Profile   *UserProfileDTO `json:"profile,omitempty"`
	Status    string          `json:"status" example:"active"`
	CreatedAt time.Time       `json:"created_at"`
}

// RoleBriefDTO is a brief role
type RoleBriefDTO struct {
	ID    string `json:"id" example:"uuid"`
	Title string `json:"title" example:"Editor"`
	Slug  string `json:"slug" example:"editor"`
}

// InviteMemberResponse is the response for inviting a member
type InviteMemberResponse struct {
	UserID  string `json:"user_id" example:"uuid"`
	Message string `json:"message" example:"member_invited"`
}

// UserRoleResponse represents a role
type UserRoleResponse struct {
	ID        string    `json:"id" example:"uuid"`
	TenantID  *string   `json:"tenant_id"`
	Title     string    `json:"title" example:"Editor"`
	Slug      string    `json:"slug" example:"editor"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RoleDetailResponse is a role with permissions
type RoleDetailResponse struct {
	Role        UserRoleResponse         `json:"role"`
	Permissions []UserPermissionResponse `json:"permissions"`
}

// UserPermissionResponse represents a permission
type UserPermissionResponse struct {
	ID          string  `json:"id" example:"uuid"`
	FeatureID   *string `json:"feature_id"`
	Title       string  `json:"title" example:"Create Products"`
	Slug        string  `json:"slug" example:"prod_c"`
	Description *string `json:"description"`
}

// ProductResponse represents a product
type ProductResponse struct {
	ID          string    `json:"id" example:"uuid"`
	TenantID    string    `json:"tenant_id" example:"uuid"`
	Name        string    `json:"name" example:"Premium Widget"`
	Description *string   `json:"description" example:"A premium widget"`
	Price       float64   `json:"price" example:"29.90"`
	SKU         *string   `json:"sku" example:"WDG-001"`
	Stock       int       `json:"stock" example:"100"`
	IsActive    bool      `json:"is_active" example:"true"`
	ImageURL    *string   `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServiceResponse represents a service
type ServiceResponse struct {
	ID          string    `json:"id" example:"uuid"`
	TenantID    string    `json:"tenant_id" example:"uuid"`
	Name        string    `json:"name" example:"Consulting"`
	Description *string   `json:"description" example:"1h consulting session"`
	Price       float64   `json:"price" example:"150.00"`
	Duration    *int      `json:"duration" example:"60"`
	IsActive    bool      `json:"is_active" example:"true"`
	ImageURL    *string   `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SettingResponse represents a tenant setting
type SettingResponse struct {
	ID        string      `json:"id" example:"uuid"`
	TenantID  string      `json:"tenant_id" example:"uuid"`
	Category  string      `json:"category" example:"appearance"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// ImageResponse represents an uploaded image
type ImageResponse struct {
	ID        string `json:"id" example:"uuid"`
	Path      string `json:"path" example:"/uploads/tenant/image.jpg"`
	PublicURL string `json:"public_url" example:"http://localhost:8080/uploads/tenant/image.jpg"`
}

// UploadResponse represents a file upload response
type UploadResponse struct {
	Path      string `json:"path" example:"/uploads/tenant/file.jpg"`
	PublicURL string `json:"public_url" example:"http://localhost:8080/uploads/tenant/file.jpg"`
}

// CreateIDResponse is a response containing a created resource ID
type CreateIDResponse struct {
	ID string `json:"id" example:"uuid"`
}

// AppUserListDTO represents an app user in backoffice lists
type AppUserListDTO struct {
	ID        string    `json:"id" example:"uuid"`
	Email     string    `json:"email" example:"appuser@example.com"`
	Name      string    `json:"name" example:"App User"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// ─── Request DTOs (Tenant API) ───────────────────────────

// SubscribeRequest is the request for public subscription
type SubscribeRequest struct {
	Name         string  `json:"name" binding:"required" example:"John Doe"`
	Email        string  `json:"email" binding:"required,email" example:"john@example.com"`
	Password     string  `json:"password" binding:"required,min=6" example:"secret123"`
	PlanID       string  `json:"plan_id" binding:"required" example:"uuid"`
	BillingCycle string  `json:"billing_cycle" binding:"required" example:"monthly"`
	PromoCode    *string `json:"promo_code" example:"LAUNCH50"`
	IsCompany    bool    `json:"is_company" example:"true"`
	CompanyName  *string `json:"company_name" example:"My Company"`
	Subdomain    string  `json:"subdomain" binding:"required" example:"mycompany"`
}

// UserLoginRequest is the request for backoffice login
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

// ChangePasswordRequest is the request for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"old_password"`
	NewPassword     string `json:"new_password" binding:"required,min=6" example:"new_password"`
}

// SelectTenantRequest is the request for selecting a tenant
type SelectTenantRequest struct {
	TenantID string `json:"tenant_id" binding:"required" example:"uuid"`
}

// UpdateUserProfileRequest is the request for updating user profile
type UpdateUserProfileRequest struct {
	FullName *string `json:"full_name" example:"John Doe"`
	About    *string `json:"about" example:"About me"`
}

// UpdateTenantProfileRequest is the request for updating tenant profile
type UpdateTenantProfileRequest struct {
	About          *string     `json:"about" example:"About our company"`
	CustomSettings interface{} `json:"custom_settings"`
}

// InviteMemberRequest is the request for inviting a member
type InviteMemberRequest struct {
	Email  string  `json:"email" binding:"required,email" example:"member@example.com"`
	Name   string  `json:"name" binding:"required" example:"Jane Doe"`
	RoleID *string `json:"role_id" example:"uuid"`
}

// UpdateMemberRoleRequest is the request for updating member role
type UpdateMemberRoleRequest struct {
	RoleID string `json:"role_id" binding:"required" example:"uuid"`
}

// CreateRoleRequest is the request for creating a role
type CreateRoleRequest struct {
	Title string `json:"title" binding:"required" example:"Editor"`
	Slug  string `json:"slug" binding:"required" example:"editor"`
}

// UpdateRoleRequest is the request for updating a role
type UpdateRoleRequest struct {
	Title string `json:"title" binding:"required" example:"Editor"`
}

// AssignPermissionRequest is the request for assigning a permission
type AssignPermissionRequest struct {
	PermissionID string `json:"permission_id" binding:"required" example:"uuid"`
}

// CreateProductRequest is the request for creating a product
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required" example:"Premium Widget"`
	Description *string `json:"description" example:"A premium widget"`
	Price       float64 `json:"price" binding:"required" example:"29.90"`
	SKU         *string `json:"sku" example:"WDG-001"`
	Stock       int     `json:"stock" example:"100"`
	IsActive    bool    `json:"is_active" example:"true"`
}

// UpdateProductRequest is the request for updating a product
type UpdateProductRequest struct {
	Name        *string  `json:"name" example:"Updated Widget"`
	Description *string  `json:"description" example:"Updated description"`
	Price       *float64 `json:"price" example:"39.90"`
	SKU         *string  `json:"sku" example:"WDG-002"`
	Stock       *int     `json:"stock" example:"50"`
	IsActive    *bool    `json:"is_active" example:"true"`
}

// CreateServiceRequest is the request for creating a service
type CreateServiceRequest struct {
	Name        string  `json:"name" binding:"required" example:"Consulting"`
	Description *string `json:"description" example:"1h consulting session"`
	Price       float64 `json:"price" binding:"required" example:"150.00"`
	Duration    *int    `json:"duration" example:"60"`
	IsActive    bool    `json:"is_active" example:"true"`
}

// UpdateServiceRequest is the request for updating a service
type UpdateServiceRequest struct {
	Name        *string  `json:"name" example:"Updated Service"`
	Description *string  `json:"description" example:"Updated description"`
	Price       *float64 `json:"price" example:"200.00"`
	Duration    *int     `json:"duration" example:"90"`
	IsActive    *bool    `json:"is_active" example:"true"`
}

// UpsertSettingRequest is the request for upserting a setting
type UpsertSettingRequest struct {
	Data interface{} `json:"data" binding:"required"`
}

// UpdateAppUserStatusRequest is the request for updating app user status
type UpdateAppUserStatusRequest struct {
	Status string `json:"status" binding:"required" example:"blocked"`
}

// ResendVerificationRequest is the request for resending verification email
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ─── Request DTOs (App API) ───────────────────────────────

// AppRegisterRequest is the request for app user registration
type AppRegisterRequest struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"secret123"`
}

// AppLoginRequest is the request for app user login
type AppLoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

// AppForgotPasswordRequest is the request for forgot password
type AppForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// AppResetPasswordRequest is the request for reset password
type AppResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"reset-token-string"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"new_password"`
}

// AppChangePasswordRequest is the request for changing password
type AppChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"old_password"`
	NewPassword     string `json:"new_password" binding:"required,min=6" example:"new_password"`
}

// AppUpdateProfileRequest is the request for updating app user profile
type AppUpdateProfileRequest struct {
	FullName *string `json:"full_name" example:"John Doe"`
	Phone    *string `json:"phone" example:"+5511999999999"`
	Document *string `json:"document" example:"123.456.789-00"`
	Address  *string `json:"address" example:"Rua A, 123"`
	Notes    *string `json:"notes" example:"Some notes"`
}

// ─── App API ─────────────────────────────────────────────

// AppAuthResponse is the response for app user auth operations
type AppAuthResponse struct {
	Token string     `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  AppUserDTO `json:"user"`
}

// AppUserDTO is a safe app user representation
type AppUserDTO struct {
	ID      string             `json:"id" example:"uuid"`
	Email   string             `json:"email" example:"user@example.com"`
	Name    string             `json:"name" example:"John"`
	Status  string             `json:"status" example:"active"`
	Profile *AppUserProfileDTO `json:"profile,omitempty"`
}

// AppUserProfileDTO is the app user profile
type AppUserProfileDTO struct {
	FullName  *string     `json:"full_name" example:"John Doe"`
	Phone     *string     `json:"phone" example:"+5511999999999"`
	Document  *string     `json:"document" example:"123.456.789-00"`
	BirthDate *string     `json:"birth_date" example:"1990-01-01"`
	AvatarURL *string     `json:"avatar_url" example:"https://example.com/avatar.jpg"`
	Address   interface{} `json:"address"`
	Metadata  interface{} `json:"metadata"`
}
