package tenant

import (
	"time"
)

// Tenant represents a tenant in the database
type Tenant struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	URLCode      string     `json:"url_code"`
	Subdomain    string     `json:"subdomain"`
	IsCompany    bool       `json:"is_company"`
	CompanyName  *string    `json:"company_name"`
	CustomDomain *string    `json:"custom_domain"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

// TenantProfile represents a tenant's profile
type TenantProfile struct {
	TenantID       string    `json:"tenant_id"`
	About          *string   `json:"about"`
	LogoURL        *string   `json:"logo_url"`
	CustomSettings any       `json:"custom_settings"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TenantPlan represents a tenant's plan subscription
type TenantPlan struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	PlanID          string     `json:"plan_id"`
	BillingCycle    string     `json:"billing_cycle"`
	BasePrice       float64    `json:"base_price"`
	ContractedPrice float64    `json:"contracted_price"`
	PriceUpdatedAt  time.Time  `json:"price_updated_at"`
	PromotionID     *string    `json:"promotion_id"`
	PromoPrice      *float64   `json:"promo_price"`
	PromoExpiresAt  *time.Time `json:"promo_expires_at"`
	IsActive        bool       `json:"is_active"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Plan represents a plan in the database
type Plan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	PlanType    string    `json:"plan_type"`
	Price       float64   `json:"price"`
	MaxUsers    int       `json:"max_users"`
	IsMultilang bool      `json:"is_multilang"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Feature represents a feature in the database
type Feature struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Code        string    `json:"code"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Promotion represents a promotion in the database
type Promotion struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description"`
	DiscountType   string     `json:"discount_type"`
	DiscountValue  float64    `json:"discount_value"`
	DurationMonths int        `json:"duration_months"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// EffectivePrice returns the current price in effect for a tenant plan
func (tp *TenantPlan) EffectivePrice() float64 {
	if tp.PromoExpiresAt != nil && tp.PromoExpiresAt.After(time.Now()) && tp.PromoPrice != nil {
		return *tp.PromoPrice
	}
	return tp.ContractedPrice
}

// --- DTOs ---

// CreateTenantRequest is the admin request to create a tenant
type CreateTenantRequest struct {
	Name          string  `json:"name" binding:"required"`
	Subdomain     string  `json:"subdomain" binding:"required"`
	IsCompany     bool    `json:"is_company"`
	CompanyName   string  `json:"company_name"`
	PlanID        string  `json:"plan_id" binding:"required"`
	BillingCycle  string  `json:"billing_cycle" binding:"required"`
	PromotionID   *string `json:"promotion_id"`
	OwnerEmail    string  `json:"owner_email"`
	OwnerFullName string  `json:"owner_full_name"`
	OwnerPassword string  `json:"owner_password"`
}

// UpdateTenantRequest is the request to update a tenant
type UpdateTenantRequest struct {
	Name         *string `json:"name"`
	IsCompany    *bool   `json:"is_company"`
	CompanyName  *string `json:"company_name"`
	CustomDomain *string `json:"custom_domain"`
}

// UpdateTenantStatusRequest is the request to update tenant status
type UpdateTenantStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ChangePlanRequest is the request to change a tenant's plan
type ChangePlanRequest struct {
	PlanID       string  `json:"plan_id" binding:"required"`
	BillingCycle string  `json:"billing_cycle" binding:"required"`
	PromotionID  *string `json:"promotion_id"`
	Reason       string  `json:"reason"`
}

// SubscriptionRequest is the public subscription request
type SubscriptionRequest struct {
	PlanID       string  `json:"plan_id" binding:"required"`
	BillingCycle string  `json:"billing_cycle"`
	PromotionID  *string `json:"promotion_id"`
	Name         string  `json:"name" binding:"required"`
	Subdomain    string  `json:"subdomain" binding:"required"`
	IsCompany    bool    `json:"is_company"`
	CompanyName  string  `json:"company_name"`
	FullName     string  `json:"full_name" binding:"required"`
	Email        string  `json:"email" binding:"required,email"`
	Password     string  `json:"password" binding:"required,min=6"`
}

// SubscriptionResponse is the response for a public subscription
type SubscriptionResponse struct {
	Tenant       TenantDTO       `json:"tenant"`
	Subscription SubscriptionDTO `json:"subscription"`
	Token        string          `json:"token"`
	User         UserBriefDTO    `json:"user"`
}

// TenantDTO is a safe representation of a tenant
type TenantDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URLCode   string `json:"url_code"`
	Subdomain string `json:"subdomain"`
	Status    string `json:"status"`
}

// SubscriptionDTO contains plan subscription details
type SubscriptionDTO struct {
	Plan            string     `json:"plan"`
	BillingCycle    string     `json:"billing_cycle"`
	ContractedPrice float64    `json:"contracted_price"`
	PromoPrice      *float64   `json:"promo_price,omitempty"`
	PromoExpiresAt  *time.Time `json:"promo_expires_at,omitempty"`
	Promotion       string     `json:"promotion,omitempty"`
}

// UserBriefDTO is a brief user representation
type UserBriefDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// CreatePlanRequest is the request to create a plan
type CreatePlanRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description *string  `json:"description"`
	PlanType    string   `json:"plan_type" binding:"required"`
	Price       float64  `json:"price" binding:"required"`
	MaxUsers    int      `json:"max_users"`
	IsMultilang bool     `json:"is_multilang"`
	FeatureIDs  []string `json:"feature_ids"`
}

// UpdatePlanRequest is the request to update a plan
type UpdatePlanRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	MaxUsers    *int     `json:"max_users"`
	IsMultilang *bool    `json:"is_multilang"`
	IsActive    *bool    `json:"is_active"`
}

// CreateFeatureRequest is the request to create a feature
type CreateFeatureRequest struct {
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

// UpdateFeatureRequest is the request to update a feature
type UpdateFeatureRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

// CreatePromotionRequest is the request to create a promotion
type CreatePromotionRequest struct {
	Name           string     `json:"name" binding:"required"`
	Description    *string    `json:"description"`
	DiscountType   string     `json:"discount_type" binding:"required"`
	DiscountValue  float64    `json:"discount_value" binding:"required"`
	DurationMonths int        `json:"duration_months" binding:"required"`
	ValidFrom      *time.Time `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until"`
	PlanID         *string    `json:"plan_id"`
}

// UpdatePromotionRequest is the request to update a promotion
type UpdatePromotionRequest struct {
	Name           *string    `json:"name"`
	Description    *string    `json:"description"`
	DiscountType   *string    `json:"discount_type"`
	DiscountValue  *float64   `json:"discount_value"`
	DurationMonths *int       `json:"duration_months"`
	ValidFrom      *time.Time `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until"`
	IsActive       *bool      `json:"is_active"`
}

// PlanFeatureRequest adds/removes a feature from a plan
type PlanFeatureRequest struct {
	FeatureID string `json:"feature_id" binding:"required"`
}

// UpdateTenantProfileRequest is the request to update a tenant's profile
type UpdateTenantProfileRequest struct {
	About          *string `json:"about"`
	CustomSettings any     `json:"custom_settings"`
}

// TenantConfigResponse is the response for tenant config
type TenantConfigResponse struct {
	Tenant      TenantConfigDTO `json:"tenant"`
	Features    []string        `json:"features"`
	Permissions []string        `json:"permissions"`
	Plan        PlanConfigDTO   `json:"plan"`
}

// TenantConfigDTO is the tenant info in config response
type TenantConfigDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	URLCode     string  `json:"url_code"`
	CompanyName *string `json:"company_name"`
}

// PlanConfigDTO is the plan info in config response
type PlanConfigDTO struct {
	Name            string     `json:"name"`
	MaxUsers        int        `json:"max_users"`
	CurrentUsers    int        `json:"current_users"`
	AvailableSlots  int        `json:"available_slots"`
	IsMultilang     bool       `json:"is_multilang"`
	BillingCycle    string     `json:"billing_cycle"`
	ContractedPrice float64    `json:"contracted_price"`
	ActivePrice     float64    `json:"active_price"`
	PromoExpiresAt  *time.Time `json:"promo_expires_at"`
	PriceUpdatedAt  time.Time  `json:"price_updated_at"`
}
