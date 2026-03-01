package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/saas-single-db-api/internal/cache"
	"github.com/saas-single-db-api/internal/email"
	repo "github.com/saas-single-db-api/internal/repository/tenant"
	"github.com/saas-single-db-api/internal/utils"
)

type Service struct {
	repo         *repo.Repository
	cache        *cache.RedisClient
	emailService *email.Service
	jwtSecret    string
	jwtExpiry    int
}

func NewService(r *repo.Repository, c *cache.RedisClient, emailSvc *email.Service, jwtSecret string, jwtExpiry int) *Service {
	return &Service{repo: r, cache: c, emailService: emailSvc, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

// --- Subscription Flow ---

type SubscribeInput struct {
	TenantName   string
	Subdomain    string
	IsCompany    bool
	CompanyName  string
	PlanID       string
	BillingCycle string
	PromoCode    *string
	OwnerName    string
	OwnerEmail   string
	OwnerPass    string
	Language     string
}

type SubscribeResult struct {
	TenantID string
	UserID   string
	Token    string
	URLCode  string
}

func (s *Service) Subscribe(ctx context.Context, input SubscribeInput) (*SubscribeResult, error) {
	// Check for existing user with same email
	existing, _ := s.repo.GetUserByEmail(ctx, input.OwnerEmail)
	if existing != nil {
		return nil, errors.New("email_already_in_use")
	}

	// Validate plan
	plan, err := s.repo.GetPlanByID(ctx, input.PlanID)
	if err != nil || !plan.IsActive {
		return nil, errors.New("invalid_or_inactive_plan")
	}

	// Auto-generate URL code
	urlCode := utils.GenerateURLCode()

	basePrice := plan.Price
	contractedPrice := basePrice
	var promoPrice *float64
	var promoExpiresAt interface{}
	var promotionID *string

	// Apply promotion if promo_code provided (lookup by name)
	if input.PromoCode != nil && *input.PromoCode != "" {
		promo, err := s.repo.GetPromotionByName(ctx, *input.PromoCode)
		if err != nil {
			return nil, errors.New("invalid_promo_code")
		}
		promoID := promo.ID
		promotionID = &promoID
		var finalPrice float64
		if promo.DiscountType == "percent" {
			finalPrice = basePrice * (1 - promo.DiscountValue/100)
		} else {
			finalPrice = basePrice - promo.DiscountValue
		}
		if finalPrice < 0 {
			finalPrice = 0
		}
		promoPrice = &finalPrice
		contractedPrice = finalPrice
		if promo.DurationMonths > 0 {
			expires := time.Now().AddDate(0, promo.DurationMonths, 0)
			promoExpiresAt = expires
		}
	}

	hashPass, err := utils.HashPassword(input.OwnerPass)
	if err != nil {
		return nil, err
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Create tenant
	tenantID, err := s.repo.CreateTenant(ctx, tx, input.TenantName, urlCode, input.Subdomain, input.IsCompany, input.CompanyName)
	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	// 2. Create tenant profile
	if err := s.repo.CreateTenantProfile(ctx, tx, tenantID); err != nil {
		return nil, fmt.Errorf("create tenant profile: %w", err)
	}

	// 3. Create tenant plan
	if err := s.repo.CreateTenantPlan(ctx, tx, tenantID, input.PlanID, input.BillingCycle, basePrice, contractedPrice, promotionID, promoPrice, promoExpiresAt); err != nil {
		return nil, fmt.Errorf("create tenant plan: %w", err)
	}

	// 4. Copy global user roles to tenant
	if err := s.repo.CopyGlobalRolesToTenant(ctx, tx, tenantID); err != nil {
		return nil, fmt.Errorf("copy roles: %w", err)
	}

	// 5. Create owner user
	userID, err := s.repo.CreateUser(ctx, tx, input.OwnerName, input.OwnerEmail, hashPass, urlCode)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// 6. Create user profile
	if err := s.repo.CreateUserProfile(ctx, tx, userID, input.OwnerName); err != nil {
		return nil, fmt.Errorf("create user profile: %w", err)
	}

	// 7. Get owner role and assign
	ownerRoleID, err := s.repo.GetTenantRoleBySlugTx(ctx, tx, tenantID, "owner")
	if err != nil {
		return nil, fmt.Errorf("get owner role: %w", err)
	}

	if err := s.repo.CreateTenantMember(ctx, tx, userID, tenantID, &ownerRoleID, true); err != nil {
		return nil, fmt.Errorf("create tenant member: %w", err)
	}

	// 8. Create email verification token
	verifyToken := utils.GenerateVerificationToken()
	if err := s.repo.CreateVerificationToken(ctx, tx, userID, verifyToken); err != nil {
		return nil, fmt.Errorf("create verification token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// 9a. Save tenant language setting
	if input.Language != "" {
		_ = s.repo.UpdateLanguage(ctx, tenantID, input.Language)
	}

	// 9. Send welcome + verification email (async, don't block on failure)
	if s.emailService != nil {
		go func() {
			verifyURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.emailService.BaseURL(), verifyToken)
			vars := map[string]string{
				"user_name":        input.OwnerName,
				"tenant_name":      input.TenantName,
				"verification_url": verifyURL,
				"url_code":         urlCode,
			}
			_ = s.emailService.SendWelcomeVerification(context.Background(), input.OwnerEmail, vars)
		}()
	}

	// Generate token
	token, err := utils.GenerateUserToken(userID, tenantID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &SubscribeResult{
		TenantID: tenantID,
		UserID:   userID,
		Token:    token,
		URLCode:  urlCode,
	}, nil
}

// VerifyEmail verifies a user's email using the verification token
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	vt, err := s.repo.GetVerificationToken(ctx, token)
	if err != nil {
		return errors.New("invalid_verification_token")
	}

	if err := s.repo.SetEmailVerified(ctx, vt.UserID); err != nil {
		return fmt.Errorf("set email verified: %w", err)
	}

	if err := s.repo.MarkTokenUsed(ctx, vt.ID); err != nil {
		return fmt.Errorf("mark token used: %w", err)
	}

	// Send confirmation email
	if s.emailService != nil {
		go func() {
			user, err := s.repo.GetUserByID(context.Background(), vt.UserID)
			if err == nil {
				vars := map[string]string{
					"user_name": user.Name,
				}
				_ = s.emailService.SendEmailVerified(context.Background(), user.Email, vars)
			}
		}()
	}

	return nil
}

// ResendVerification creates a new verification token and sends the email
func (s *Service) ResendVerification(ctx context.Context, userID string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("user_not_found")
	}

	if user.EmailVerifiedAt != nil {
		return errors.New("email_already_verified")
	}

	verifyToken := utils.GenerateVerificationToken()
	if err := s.repo.CreateVerificationTokenNonTx(ctx, userID, verifyToken); err != nil {
		return fmt.Errorf("create verification token: %w", err)
	}

	if s.emailService != nil {
		verifyURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.emailService.BaseURL(), verifyToken)
		vars := map[string]string{
			"user_name":        user.Name,
			"verification_url": verifyURL,
		}
		return s.emailService.SendWelcomeVerification(ctx, user.Email, vars)
	}

	return nil
}

// --- Auth ---

type LoginResult struct {
	Token             string
	Name              string
	Email             string
	CurrentTenantCode string
	Tenants           []repo.TenantBriefExported
	Language          string
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid_credentials")
	}
	if user.Status != "active" {
		return nil, errors.New("account_not_active")
	}
	if !utils.CheckPassword(password, user.HashPass) {
		return nil, errors.New("invalid_credentials")
	}

	tenants, _ := s.repo.GetUserTenants(ctx, user.ID)
	tenantList := make([]repo.TenantBriefExported, len(tenants))
	for i, t := range tenants {
		tenantList[i] = repo.TenantBriefExported{
			ID: t.ID, URLCode: t.URLCode, Name: t.Name, IsOwner: t.IsOwner,
		}
	}

	// Use first tenant as default if no last_tenant
	urlCode := ""
	tenantID := ""
	if user.LastTenantURLCode != nil && *user.LastTenantURLCode != "" {
		urlCode = *user.LastTenantURLCode
		tid, err := s.repo.GetTenantByURLCode(ctx, urlCode)
		if err == nil {
			tenantID = tid
		}
	}
	if tenantID == "" && len(tenants) > 0 {
		urlCode = tenants[0].URLCode
		tenantID = tenants[0].ID
	}

	token, err := utils.GenerateUserToken(user.ID, tenantID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	// Fetch language for the current tenant
	language := s.repo.GetLanguage(ctx, tenantID)

	return &LoginResult{
		Token:             token,
		Name:              user.Name,
		Email:             user.Email,
		CurrentTenantCode: urlCode,
		Tenants:           tenantList,
		Language:          language,
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID string) (map[string]interface{}, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user_not_found")
	}
	profile, _ := s.repo.GetUserProfile(ctx, userID)
	tenants, _ := s.repo.GetUserTenants(ctx, userID)

	result := map[string]interface{}{
		"id":     user.ID,
		"name":   user.Name,
		"email":  user.Email,
		"status": user.Status,
	}
	if profile != nil {
		result["profile"] = map[string]interface{}{
			"full_name":  profile.FullName,
			"about":      profile.About,
			"avatar_url": profile.AvatarURL,
		}
	}
	result["tenants"] = tenants
	return result, nil
}

func (s *Service) HasPermission(ctx context.Context, userID, tenantID, permSlug string) bool {
	perms, err := s.repo.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return false
	}
	for _, p := range perms {
		if p == permSlug {
			return true
		}
	}
	return false
}

func (s *Service) IsOwner(ctx context.Context, userID, tenantID string) bool {
	member, err := s.repo.GetMember(ctx, tenantID, userID)
	if err != nil {
		return false
	}
	return member.IsOwner
}

// --- Members ---

func (s *Service) CanAddMember(ctx context.Context, tenantID string) (bool, int, int, error) {
	plan, err := s.repo.GetActiveTenantPlan(ctx, tenantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, 0, 0, errors.New("no_active_plan")
		}
		return false, 0, 0, err
	}

	count, err := s.repo.CountTenantMembers(ctx, tenantID)
	if err != nil {
		return false, 0, 0, err
	}

	canAdd := count < plan.MaxUsers
	return canAdd, plan.MaxUsers, count, nil
}

func (s *Service) InviteMember(ctx context.Context, tenantID, email, name, password, roleSlug string) (string, error) {
	canAdd, _, _, err := s.CanAddMember(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if !canAdd {
		return "", fmt.Errorf("max_users_reached")
	}

	// Check if user exists
	user, _ := s.repo.GetUserByEmail(ctx, email)
	var userID string

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	if user != nil {
		userID = user.ID
		// Check if already a member
		if s.repo.IsMember(ctx, userID, tenantID) {
			return "", errors.New("user_already_member")
		}
	} else {
		hashPass, err := utils.HashPassword(password)
		if err != nil {
			return "", err
		}
		tenant, err := s.repo.GetTenantByID(ctx, tenantID)
		if err != nil {
			return "", err
		}
		userID, err = s.repo.CreateUser(ctx, tx, name, email, hashPass, tenant.URLCode)
		if err != nil {
			return "", fmt.Errorf("create user: %w", err)
		}
		if err := s.repo.CreateUserProfile(ctx, tx, userID, name); err != nil {
			return "", fmt.Errorf("create profile: %w", err)
		}
	}

	roleID, err := s.repo.GetTenantRoleBySlugTx(ctx, tx, tenantID, roleSlug)
	if err != nil {
		return "", fmt.Errorf("role_not_found_tenant")
	}

	if err := s.repo.CreateTenantMember(ctx, tx, userID, tenantID, &roleID, false); err != nil {
		return "", fmt.Errorf("create member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return userID, nil
}
