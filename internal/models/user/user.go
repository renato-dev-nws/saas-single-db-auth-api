package user

import (
	"time"
)

// User represents a backoffice user in the database
type User struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	HashPass          string     `json:"-"`
	LastTenantURLCode *string    `json:"last_tenant_url_code"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"-"`
}

// UserProfile represents a user's profile
type UserProfile struct {
	UserID    string    `json:"user_id"`
	FullName  *string   `json:"full_name"`
	About     *string   `json:"about"`
	AvatarURL *string   `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRole represents a role for tenant users
type UserRole struct {
	ID        string    `json:"id"`
	TenantID  *string   `json:"tenant_id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserPermission represents a permission
type UserPermission struct {
	ID          string    `json:"id"`
	FeatureID   *string   `json:"feature_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TenantMember represents membership of a user in a tenant
type TenantMember struct {
	UserID    string     `json:"user_id"`
	TenantID  string     `json:"tenant_id"`
	RoleID    *string    `json:"role_id"`
	IsOwner   bool       `json:"is_owner"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

// Product represents a product
type Product struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Price       float64   `json:"price"`
	SKU         *string   `json:"sku"`
	Stock       int       `json:"stock"`
	IsActive    bool      `json:"is_active"`
	ImageURL    *string   `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Service represents a service
type Service struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Price       float64   `json:"price"`
	Duration    *int      `json:"duration"`
	IsActive    bool      `json:"is_active"`
	ImageURL    *string   `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Setting represents a tenant's setting
type Setting struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Category  string    `json:"category"`
	Data      any       `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Image represents an uploaded image
type Image struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	OriginalName string    `json:"original_name"`
	StoragePath  string    `json:"storage_path"`
	PublicURL    string    `json:"public_url"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	Width        *int      `json:"width"`
	Height       *int      `json:"height"`
	Provider     string    `json:"provider"`
	EntityType   *string   `json:"entity_type"`
	EntityID     *string   `json:"entity_id"`
	UploadedBy   *string   `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}

// --- DTOs ---

// LoginRequest is used for backoffice user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is the response for backoffice user login
type LoginResponse struct {
	Token         string        `json:"token"`
	User          UserDTO       `json:"user"`
	CurrentTenant *TenantInfo   `json:"current_tenant"`
	Tenants       []TenantBrief `json:"tenants"`
}

// UserDTO is a safe representation of a user
type UserDTO struct {
	ID      string      `json:"id"`
	Email   string      `json:"email"`
	Name    string      `json:"name"`
	Status  string      `json:"status"`
	Profile *ProfileDTO `json:"profile,omitempty"`
}

// ProfileDTO is the profile DTO for users
type ProfileDTO struct {
	FullName  *string `json:"full_name"`
	About     *string `json:"about"`
	AvatarURL *string `json:"avatar_url"`
}

// TenantInfo includes features and permissions for a specific tenant
type TenantInfo struct {
	ID          string   `json:"id"`
	URLCode     string   `json:"url_code"`
	CompanyName *string  `json:"company_name"`
	Name        string   `json:"name"`
	Features    []string `json:"features"`
	Permissions []string `json:"permissions"`
}

// TenantBrief is a brief tenant info
type TenantBrief struct {
	ID      string `json:"id"`
	URLCode string `json:"url_code"`
	Name    string `json:"name"`
	IsOwner bool   `json:"is_owner"`
}

// UpdateProfileRequest update profile request
type UpdateProfileRequest struct {
	FullName *string `json:"full_name"`
	About    *string `json:"about"`
}

// ChangePasswordRequest change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// MemberDTO represents a member of a tenant
type MemberDTO struct {
	UserID    string      `json:"user_id"`
	Email     string      `json:"email"`
	Name      string      `json:"name"`
	IsOwner   bool        `json:"is_owner"`
	Role      *RoleBrief  `json:"role"`
	Profile   *ProfileDTO `json:"profile,omitempty"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

// RoleBrief is a brief role representation
type RoleBrief struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

// CanAddMemberResponse returns whether a member can be added
type CanAddMemberResponse struct {
	CanAdd         bool   `json:"can_add"`
	CurrentUsers   int    `json:"current_users"`
	MaxUsers       int    `json:"max_users"`
	AvailableSlots int    `json:"available_slots"`
	Reason         string `json:"reason,omitempty"`
	UpgradeHint    string `json:"upgrade_hint,omitempty"`
}

// CreateMemberRequest creates a member
type CreateMemberRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	RoleSlug string `json:"role_slug"`
}

// UpdateMemberRoleRequest updates a member's role
type UpdateMemberRoleRequest struct {
	RoleSlug string `json:"role_slug" binding:"required"`
}

// CreateRoleRequest creates a role
type CreateRoleRequest struct {
	Title string `json:"title" binding:"required"`
	Slug  string `json:"slug" binding:"required"`
}

// UpdateRoleRequest updates a role
type UpdateRoleRequest struct {
	Title *string `json:"title"`
}

// AssignPermissionRequest assigns a permission to a role
type AssignPermissionRequest struct {
	PermissionID string `json:"permission_id" binding:"required"`
}

// CreateProductRequest creates a product
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	SKU         *string `json:"sku"`
	Stock       int     `json:"stock"`
}

// UpdateProductRequest updates a product
type UpdateProductRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	SKU         *string  `json:"sku"`
	Stock       *int     `json:"stock"`
	IsActive    *bool    `json:"is_active"`
}

// CreateServiceRequest creates a service
type CreateServiceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	Duration    *int    `json:"duration"`
}

// UpdateServiceRequest updates a service
type UpdateServiceRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Duration    *int     `json:"duration"`
	IsActive    *bool    `json:"is_active"`
}

// UpdateSettingRequest updates a setting category
type UpdateSettingRequest struct {
	Data any `json:"data" binding:"required"`
}

// UpdateAppUserStatusRequest updates an app user's status
type UpdateAppUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
