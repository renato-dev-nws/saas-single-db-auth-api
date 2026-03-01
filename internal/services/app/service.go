package app

import (
	"context"
	"errors"

	repo "github.com/saas-single-db-api/internal/repository/app"
	"github.com/saas-single-db-api/internal/utils"
)

type Service struct {
	repo      *repo.Repository
	jwtSecret string
	jwtExpiry int
}

func NewService(r *repo.Repository, jwtSecret string, jwtExpiry int) *Service {
	return &Service{repo: r, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

type RegisterResult struct {
	UserID string
	Token  string
}

func (s *Service) Register(ctx context.Context, tenantID, urlCode, name, email, password string) (*RegisterResult, error) {
	existing, _ := s.repo.GetAppUserByEmail(ctx, tenantID, email)
	if existing != nil {
		return nil, errors.New("email_already_registered")
	}

	hashPass, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	userID, err := s.repo.CreateAppUser(ctx, tenantID, name, email, hashPass)
	if err != nil {
		return nil, err
	}

	_ = s.repo.CreateAppUserProfile(ctx, userID, name)

	token, err := utils.GenerateAppUserToken(userID, tenantID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &RegisterResult{UserID: userID, Token: token}, nil
}

type LoginResult struct {
	UserID string
	Token  string
	Name   string
	Email  string
}

func (s *Service) Login(ctx context.Context, tenantID, urlCode, email, password string) (*LoginResult, error) {
	user, err := s.repo.GetAppUserByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, errors.New("invalid_credentials")
	}
	if user.Status != "active" {
		return nil, errors.New("account_not_active")
	}
	if !utils.CheckPassword(password, user.HashPass) {
		return nil, errors.New("invalid_credentials")
	}

	token, err := utils.GenerateAppUserToken(user.ID, tenantID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID: user.ID,
		Token:  token,
		Name:   user.Name,
		Email:  user.Email,
	}, nil
}

func (s *Service) GetMe(ctx context.Context, tenantID, userID string) (map[string]interface{}, error) {
	user, err := s.repo.GetAppUserByID(ctx, tenantID, userID)
	if err != nil {
		return nil, errors.New("user_not_found")
	}

	profile, _ := s.repo.GetAppUserProfile(ctx, userID)

	result := map[string]interface{}{
		"id":     user.ID,
		"name":   user.Name,
		"email":  user.Email,
		"status": user.Status,
	}
	if profile != nil {
		result["profile"] = map[string]interface{}{
			"full_name":  profile.FullName,
			"phone":      profile.Phone,
			"document":   profile.Document,
			"avatar_url": profile.AvatarURL,
			"address":    profile.Address,
			"notes":      profile.Notes,
		}
	}
	return result, nil
}

func (s *Service) ChangePassword(ctx context.Context, tenantID, userID, currentPass, newPass string) error {
	user, err := s.repo.GetAppUserByID(ctx, tenantID, userID)
	if err != nil {
		return errors.New("user_not_found")
	}
	if !utils.CheckPassword(currentPass, user.HashPass) {
		return errors.New("current_password_incorrect")
	}
	hashPass, err := utils.HashPassword(newPass)
	if err != nil {
		return err
	}
	return s.repo.UpdateAppUserPassword(ctx, tenantID, userID, hashPass)
}

func (s *Service) ForgotPassword(ctx context.Context, tenantID, email string) error {
	_, err := s.repo.GetAppUserByEmail(ctx, tenantID, email)
	if err != nil {
		// Don't reveal if email exists
		return nil
	}
	// TODO: Send password reset email
	// For now, just return success
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, tenantID, token, newPass string) error {
	// TODO: Validate token and get user ID
	// For now, return not implemented
	return errors.New("password_reset_not_implemented")
}
