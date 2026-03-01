package admin

import (
	"time"
)

// SystemAdminUser represents a system admin user in the database
type SystemAdminUser struct {
	ID        string     `json:"id"`
	Name      *string    `json:"name"`
	Email     string     `json:"email"`
	HashPass  string     `json:"-"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

// SystemAdminProfile represents an admin user's profile
type SystemAdminProfile struct {
	AdminUserID string    `json:"admin_user_id"`
	FullName    *string   `json:"full_name"`
	Title       *string   `json:"title"`
	Bio         *string   `json:"bio"`
	AvatarURL   *string   `json:"avatar_url"`
	SocialLinks any       `json:"social_links"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SystemAdminRole represents an admin role
type SystemAdminRole struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SystemAdminPermission represents an admin permission
type SystemAdminPermission struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- DTOs ---

// LoginRequest is the request body for admin login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is the response for admin login
type LoginResponse struct {
	Token string       `json:"token"`
	Admin AdminUserDTO `json:"admin"`
}

// AdminUserDTO is a safe representation of an admin user (no password hash)
type AdminUserDTO struct {
	ID      string      `json:"id"`
	Email   string      `json:"email"`
	Name    *string     `json:"name"`
	Status  string      `json:"status"`
	Profile *ProfileDTO `json:"profile,omitempty"`
	Roles   []RoleDTO   `json:"roles,omitempty"`
}

// ProfileDTO is the profile DTO
type ProfileDTO struct {
	FullName    *string `json:"full_name"`
	Title       *string `json:"title"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
	SocialLinks any     `json:"social_links"`
}

// RoleDTO is the role DTO
type RoleDTO struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
}

// CreateAdminRequest is the request body for creating an admin
type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name"`
	RoleSlug string `json:"role_slug"`
}

// UpdateAdminRequest is the request body for updating an admin
type UpdateAdminRequest struct {
	Name   *string `json:"name"`
	Email  *string `json:"email"`
	Status *string `json:"status"`
}

// UpdateProfileRequest is the request body for updating a profile
type UpdateProfileRequest struct {
	FullName    *string `json:"full_name"`
	Title       *string `json:"title"`
	Bio         *string `json:"bio"`
	SocialLinks any     `json:"social_links"`
}

// ChangePasswordRequest is the request body for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// CreateRoleRequest is the request body for creating a role
type CreateRoleRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description *string `json:"description"`
}

// UpdateRoleRequest is the request body for updating a role
type UpdateRoleRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

// AssignRoleRequest is the request body for assigning a role
type AssignRoleRequest struct {
	RoleID int `json:"role_id" binding:"required"`
}
