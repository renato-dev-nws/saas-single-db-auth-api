package app

import (
	"time"
)

// TenantAppUser represents a client/end user of a tenant
type TenantAppUser struct {
	ID        string     `json:"id"`
	TenantID  string     `json:"tenant_id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	HashPass  string     `json:"-"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

// TenantAppUserProfile represents an app user's profile
type TenantAppUserProfile struct {
	AppUserID string    `json:"app_user_id"`
	FullName  *string   `json:"full_name"`
	Phone     *string   `json:"phone"`
	Document  *string   `json:"document"`
	BirthDate *string   `json:"birth_date"`
	AvatarURL *string   `json:"avatar_url"`
	Address   any       `json:"address"`
	Metadata  any       `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- DTOs ---

// RegisterRequest is the request for app user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Phone    string `json:"phone"`
}

// LoginRequest is the request for app user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is the response for auth operations
type AuthResponse struct {
	Token string     `json:"token"`
	User  AppUserDTO `json:"user"`
}

// AppUserDTO is a safe representation of an app user
type AppUserDTO struct {
	ID      string      `json:"id"`
	Email   string      `json:"email"`
	Name    string      `json:"name"`
	Status  string      `json:"status"`
	Profile *ProfileDTO `json:"profile,omitempty"`
}

// ProfileDTO is the profile DTO for app users
type ProfileDTO struct {
	FullName  *string `json:"full_name"`
	Phone     *string `json:"phone"`
	Document  *string `json:"document"`
	BirthDate *string `json:"birth_date"`
	AvatarURL *string `json:"avatar_url"`
	Address   any     `json:"address"`
	Metadata  any     `json:"metadata"`
}

// UpdateProfileRequest updates an app user's profile
type UpdateProfileRequest struct {
	FullName  *string `json:"full_name"`
	Phone     *string `json:"phone"`
	Document  *string `json:"document"`
	BirthDate *string `json:"birth_date"`
	Address   any     `json:"address"`
	Metadata  any     `json:"metadata"`
}

// ChangePasswordRequest change password request for app user
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ForgotPasswordRequest forgot password
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest reset password
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
