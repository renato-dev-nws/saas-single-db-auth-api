package admin

import (
	"context"
	"errors"
	"slices"

	repo "github.com/saas-single-db-api/internal/repository/admin"
	"github.com/saas-single-db-api/internal/utils"
)

type Service struct {
	repo      *repo.Repository
	jwtSecret string
	jwtExpiry int
}

func NewService(repo *repo.Repository, jwtSecret string, jwtExpiry int) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

func (s *Service) Login(ctx context.Context, email, password string) (string, interface{}, error) {
	admin, err := s.repo.GetAdminByEmail(ctx, email)
	if err != nil {
		return "", nil, errors.New("invalid_credentials")
	}
	if admin.Status != "active" {
		return "", nil, errors.New("account_suspended")
	}
	if !utils.CheckPassword(password, admin.HashPass) {
		return "", nil, errors.New("invalid_credentials")
	}

	token, err := utils.GenerateAdminToken(admin.ID, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return "", nil, err
	}

	profile, _ := s.repo.GetProfile(ctx, admin.ID)

	result := map[string]interface{}{
		"id":    admin.ID,
		"email": admin.Email,
		"name":  admin.Name,
	}
	if profile != nil {
		result["profile"] = map[string]interface{}{
			"full_name": profile.FullName,
		}
	}

	return token, result, nil
}

func (s *Service) GetMe(ctx context.Context, adminID string) (interface{}, error) {
	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	profile, _ := s.repo.GetProfile(ctx, adminID)
	roles, _ := s.repo.GetAdminRoles(ctx, adminID)
	permissions, _ := s.repo.GetAdminPermissions(ctx, adminID)

	result := map[string]interface{}{
		"id":          admin.ID,
		"email":       admin.Email,
		"name":        admin.Name,
		"status":      admin.Status,
		"roles":       roles,
		"permissions": permissions,
	}
	if profile != nil {
		result["profile"] = profile
	}
	return result, nil
}

func (s *Service) HasPermission(ctx context.Context, adminID, permission string) bool {
	perms, err := s.repo.GetAdminPermissions(ctx, adminID)
	if err != nil {
		return false
	}
	return slices.Contains(perms, permission)
}

func (s *Service) Repo() *repo.Repository {
	return s.repo
}

func (s *Service) JWTSecret() string {
	return s.jwtSecret
}

func (s *Service) JWTExpiry() int {
	return s.jwtExpiry
}
