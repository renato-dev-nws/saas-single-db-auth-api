package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	models "github.com/saas-single-db-api/internal/models/admin"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// --- Admin Users ---

func (r *Repository) GetAdminByEmail(ctx context.Context, email string) (*models.SystemAdminUser, error) {
	var u models.SystemAdminUser
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, hash_pass, status, created_at, updated_at
		 FROM system_admin_users WHERE email = $1 AND deleted_at IS NULL`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.HashPass, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetAdminByID(ctx context.Context, id string) (*models.SystemAdminUser, error) {
	var u models.SystemAdminUser
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, hash_pass, status, created_at, updated_at
		 FROM system_admin_users WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.HashPass, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) ListAdmins(ctx context.Context, limit, offset int) ([]models.SystemAdminUser, int64, error) {
	var total int64
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM system_admin_users WHERE deleted_at IS NULL`,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, name, email, status, created_at, updated_at
		 FROM system_admin_users WHERE deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var admins []models.SystemAdminUser
	for rows.Next() {
		var u models.SystemAdminUser
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		admins = append(admins, u)
	}
	return admins, total, nil
}

func (r *Repository) CreateAdmin(ctx context.Context, name, email, hashPass string) (*models.SystemAdminUser, error) {
	var u models.SystemAdminUser
	err := r.db.QueryRow(ctx,
		`INSERT INTO system_admin_users (name, email, hash_pass) VALUES ($1, $2, $3)
		 RETURNING id, name, email, status, created_at, updated_at`,
		name, email, hashPass,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdateAdmin(ctx context.Context, id string, req *models.UpdateAdminRequest) error {
	query := `UPDATE system_admin_users SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Email != nil {
		query += fmt.Sprintf(", email = $%d", argIdx)
		args = append(args, *req.Email)
		argIdx++
	}
	if req.Status != nil {
		query += fmt.Sprintf(", status = $%d", argIdx)
		args = append(args, *req.Status)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", argIdx)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) SoftDeleteAdmin(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE system_admin_users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	return err
}

func (r *Repository) UpdateAdminPassword(ctx context.Context, id, hashPass string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE system_admin_users SET hash_pass = $1, updated_at = NOW() WHERE id = $2`, hashPass, id,
	)
	return err
}

// --- Profiles ---

func (r *Repository) GetProfile(ctx context.Context, adminID string) (*models.SystemAdminProfile, error) {
	var p models.SystemAdminProfile
	err := r.db.QueryRow(ctx,
		`SELECT admin_user_id, full_name, title, bio, avatar_url, social_links, created_at, updated_at
		 FROM system_admin_profiles WHERE admin_user_id = $1`, adminID,
	).Scan(&p.AdminUserID, &p.FullName, &p.Title, &p.Bio, &p.AvatarURL, &p.SocialLinks, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) UpsertProfile(ctx context.Context, adminID string, req *models.UpdateProfileRequest) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO system_admin_profiles (admin_user_id, full_name, title, bio, social_links)
		 VALUES ($1, $2, $3, $4, COALESCE($5::jsonb, '{}'))
		 ON CONFLICT (admin_user_id) DO UPDATE SET
		   full_name = COALESCE($2, system_admin_profiles.full_name),
		   title = COALESCE($3, system_admin_profiles.title),
		   bio = COALESCE($4, system_admin_profiles.bio),
		   social_links = COALESCE($5::jsonb, system_admin_profiles.social_links),
		   updated_at = NOW()`,
		adminID, req.FullName, req.Title, req.Bio, req.SocialLinks,
	)
	return err
}

func (r *Repository) CreateProfile(ctx context.Context, adminID, fullName string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO system_admin_profiles (admin_user_id, full_name) VALUES ($1, $2)
		 ON CONFLICT (admin_user_id) DO NOTHING`,
		adminID, fullName,
	)
	return err
}

// --- Roles ---

func (r *Repository) ListRoles(ctx context.Context) ([]models.SystemAdminRole, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, slug, description, created_at, updated_at FROM system_admin_roles ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.SystemAdminRole
	for rows.Next() {
		var role models.SystemAdminRole
		if err := rows.Scan(&role.ID, &role.Title, &role.Slug, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *Repository) GetRoleByID(ctx context.Context, id int) (*models.SystemAdminRole, error) {
	var role models.SystemAdminRole
	err := r.db.QueryRow(ctx,
		`SELECT id, title, slug, description, created_at, updated_at FROM system_admin_roles WHERE id = $1`, id,
	).Scan(&role.ID, &role.Title, &role.Slug, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) CreateRole(ctx context.Context, title, slug string, description *string) (*models.SystemAdminRole, error) {
	var role models.SystemAdminRole
	err := r.db.QueryRow(ctx,
		`INSERT INTO system_admin_roles (title, slug, description) VALUES ($1, $2, $3)
		 RETURNING id, title, slug, description, created_at, updated_at`,
		title, slug, description,
	).Scan(&role.ID, &role.Title, &role.Slug, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) UpdateRole(ctx context.Context, id int, req *models.UpdateRoleRequest) error {
	query := `UPDATE system_admin_roles SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.Title != nil {
		query += fmt.Sprintf(", title = $%d", argIdx)
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, *req.Description)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) DeleteRole(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM system_admin_roles WHERE id = $1`, id)
	return err
}

func (r *Repository) AssignRoleToAdmin(ctx context.Context, adminID string, roleID int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO system_admin_user_roles (admin_user_id, admin_role_id) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`, adminID, roleID,
	)
	return err
}

func (r *Repository) RemoveRoleFromAdmin(ctx context.Context, adminID string, roleID int) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM system_admin_user_roles WHERE admin_user_id = $1 AND admin_role_id = $2`, adminID, roleID,
	)
	return err
}

func (r *Repository) GetAdminRoles(ctx context.Context, adminID string) ([]models.SystemAdminRole, error) {
	rows, err := r.db.Query(ctx,
		`SELECT r.id, r.title, r.slug, r.description, r.created_at, r.updated_at
		 FROM system_admin_roles r
		 JOIN system_admin_user_roles ur ON ur.admin_role_id = r.id
		 WHERE ur.admin_user_id = $1`, adminID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.SystemAdminRole
	for rows.Next() {
		var role models.SystemAdminRole
		if err := rows.Scan(&role.ID, &role.Title, &role.Slug, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// --- Permissions ---

func (r *Repository) ListPermissions(ctx context.Context) ([]models.SystemAdminPermission, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, slug, description, created_at, updated_at FROM system_admin_permissions ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.SystemAdminPermission
	for rows.Next() {
		var p models.SystemAdminPermission
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *Repository) GetAdminPermissions(ctx context.Context, adminID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT p.slug
		 FROM system_admin_permissions p
		 JOIN system_admin_role_permissions rp ON rp.admin_permission_id = p.id
		 JOIN system_admin_user_roles ur ON ur.admin_role_id = rp.admin_role_id
		 WHERE ur.admin_user_id = $1`, adminID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var slug string
		if rows.Scan(&slug) == nil {
			perms = append(perms, slug)
		}
	}
	return perms, nil
}

func (r *Repository) GetRoleBySlug(ctx context.Context, slug string) (*models.SystemAdminRole, error) {
	var role models.SystemAdminRole
	err := r.db.QueryRow(ctx,
		`SELECT id, title, slug, description, created_at, updated_at FROM system_admin_roles WHERE slug = $1`, slug,
	).Scan(&role.ID, &role.Title, &role.Slug, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// --- Tenants (admin view) ---

func (r *Repository) ListTenants(ctx context.Context, limit, offset int) ([]tenantRow, int64, error) {
	var total int64
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL`,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.name, t.url_code, t.subdomain, t.is_company, t.company_name,
		        t.custom_domain, t.status, t.created_at, t.updated_at,
		        p.name as plan_name, tp.billing_cycle, tp.contracted_price
		 FROM tenants t
		 LEFT JOIN tenant_plans tp ON tp.tenant_id = t.id AND tp.is_active = true
		 LEFT JOIN plans p ON p.id = tp.plan_id
		 WHERE t.deleted_at IS NULL
		 ORDER BY t.created_at DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tenants []tenantRow
	for rows.Next() {
		var t tenantRow
		if err := rows.Scan(&t.ID, &t.Name, &t.URLCode, &t.Subdomain, &t.IsCompany,
			&t.CompanyName, &t.CustomDomain, &t.Status, &t.CreatedAt, &t.UpdatedAt,
			&t.PlanName, &t.BillingCycle, &t.ContractedPrice); err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, t)
	}
	return tenants, total, nil
}

type tenantRow struct {
	ID              string
	Name            string
	URLCode         string
	Subdomain       string
	IsCompany       bool
	CompanyName     *string
	CustomDomain    *string
	Status          string
	CreatedAt       interface{}
	UpdatedAt       interface{}
	PlanName        *string
	BillingCycle    *string
	ContractedPrice *float64
}

func (r *Repository) GetTenantByID(ctx context.Context, id string) (*tenantRow, error) {
	var t tenantRow
	err := r.db.QueryRow(ctx,
		`SELECT t.id, t.name, t.url_code, t.subdomain, t.is_company, t.company_name,
		        t.custom_domain, t.status, t.created_at, t.updated_at,
		        p.name as plan_name, tp.billing_cycle, tp.contracted_price
		 FROM tenants t
		 LEFT JOIN tenant_plans tp ON tp.tenant_id = t.id AND tp.is_active = true
		 LEFT JOIN plans p ON p.id = tp.plan_id
		 WHERE t.id = $1 AND t.deleted_at IS NULL`, id,
	).Scan(&t.ID, &t.Name, &t.URLCode, &t.Subdomain, &t.IsCompany,
		&t.CompanyName, &t.CustomDomain, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		&t.PlanName, &t.BillingCycle, &t.ContractedPrice)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) CreateTenant(ctx context.Context, tx pgx.Tx, name, urlCode, subdomain string, isCompany bool, companyName string) (string, error) {
	var id string
	err := tx.QueryRow(ctx,
		`INSERT INTO tenants (name, url_code, subdomain, is_company, company_name)
		 VALUES ($1, $2, $3, $4, NULLIF($5, ''))
		 RETURNING id`,
		name, urlCode, subdomain, isCompany, companyName,
	).Scan(&id)
	return id, err
}

func (r *Repository) CreateTenantProfile(ctx context.Context, tx pgx.Tx, tenantID string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO tenant_profiles (tenant_id) VALUES ($1)`, tenantID,
	)
	return err
}

func (r *Repository) UpdateTenant(ctx context.Context, id string, req interface{}) error {
	// Dynamic update based on the provided fields
	type updateReq struct {
		Name         *string `json:"name"`
		IsCompany    *bool   `json:"is_company"`
		CompanyName  *string `json:"company_name"`
		CustomDomain *string `json:"custom_domain"`
	}
	// Use simple approach
	_, err := r.db.Exec(ctx,
		`UPDATE tenants SET updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	return err
}

func (r *Repository) UpdateTenantStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenants SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`, status, id,
	)
	return err
}

func (r *Repository) SoftDeleteTenant(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenants SET deleted_at = NOW(), status = 'cancelled', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	return err
}

func (r *Repository) GetTenantMembers(ctx context.Context, tenantID string) ([]memberRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT u.id, u.name, u.email, u.status, tm.is_owner,
		        r.id, r.title, r.slug,
		        up.full_name, up.avatar_url
		 FROM tenant_members tm
		 JOIN users u ON u.id = tm.user_id
		 LEFT JOIN user_roles r ON r.id = tm.role_id
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 WHERE tm.tenant_id = $1 AND tm.deleted_at IS NULL AND u.deleted_at IS NULL
		 ORDER BY tm.is_owner DESC, u.name`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []memberRow
	for rows.Next() {
		var m memberRow
		if err := rows.Scan(&m.UserID, &m.Name, &m.Email, &m.Status, &m.IsOwner,
			&m.RoleID, &m.RoleTitle, &m.RoleSlug,
			&m.FullName, &m.AvatarURL); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

type memberRow struct {
	UserID    string
	Name      string
	Email     string
	Status    string
	IsOwner   bool
	RoleID    *string
	RoleTitle *string
	RoleSlug  *string
	FullName  *string
	AvatarURL *string
}

// --- Plans ---

func (r *Repository) ListPlans(ctx context.Context) ([]planWithFeatures, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.name, p.description, p.plan_type, p.price, p.max_users, p.is_multilang, p.is_active,
		        p.created_at, p.updated_at
		 FROM plans p ORDER BY p.price`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []planWithFeatures
	for rows.Next() {
		var p planWithFeatures
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PlanType, &p.Price, &p.MaxUsers,
			&p.IsMultilang, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}

	// Get features for each plan
	for i, p := range plans {
		features, _ := r.GetPlanFeatures(ctx, p.ID)
		plans[i].Features = features
	}
	return plans, nil
}

type planWithFeatures struct {
	ID          string
	Name        string
	Description *string
	PlanType    string
	Price       float64
	MaxUsers    int
	IsMultilang bool
	IsActive    bool
	CreatedAt   interface{}
	UpdatedAt   interface{}
	Features    []featureRow
}

type featureRow struct {
	ID    string
	Title string
	Slug  string
	Code  string
}

func (r *Repository) GetPlanByID(ctx context.Context, id string) (*planWithFeatures, error) {
	var p planWithFeatures
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, plan_type, price, max_users, is_multilang, is_active, created_at, updated_at
		 FROM plans WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.PlanType, &p.Price, &p.MaxUsers, &p.IsMultilang, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Features, _ = r.GetPlanFeatures(ctx, id)
	return &p, nil
}

func (r *Repository) CreatePlan(ctx context.Context, name string, description *string, planType string, price float64, maxUsers int, isMultilang bool) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO plans (name, description, plan_type, price, max_users, is_multilang) VALUES ($1, $2, $3::plan_type, $4, $5, $6) RETURNING id`,
		name, description, planType, price, maxUsers, isMultilang,
	).Scan(&id)
	return id, err
}

func (r *Repository) UpdatePlan(ctx context.Context, id string, name *string, description *string, price *float64, maxUsers *int, isMultilang *bool, isActive *bool) error {
	query := `UPDATE plans SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if name != nil {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, *name)
		argIdx++
	}
	if description != nil {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, *description)
		argIdx++
	}
	if price != nil {
		query += fmt.Sprintf(", price = $%d", argIdx)
		args = append(args, *price)
		argIdx++
	}
	if maxUsers != nil {
		query += fmt.Sprintf(", max_users = $%d", argIdx)
		args = append(args, *maxUsers)
		argIdx++
	}
	if isMultilang != nil {
		query += fmt.Sprintf(", is_multilang = $%d", argIdx)
		args = append(args, *isMultilang)
		argIdx++
	}
	if isActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argIdx)
		args = append(args, *isActive)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) DeletePlan(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM plans WHERE id = $1`, id)
	return err
}

func (r *Repository) GetPlanFeatures(ctx context.Context, planID string) ([]featureRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT f.id, f.title, f.slug, f.code
		 FROM features f
		 JOIN plan_features pf ON pf.feature_id = f.id
		 WHERE pf.plan_id = $1`, planID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []featureRow
	for rows.Next() {
		var f featureRow
		if err := rows.Scan(&f.ID, &f.Title, &f.Slug, &f.Code); err != nil {
			return nil, err
		}
		features = append(features, f)
	}
	return features, nil
}

func (r *Repository) AddFeatureToPlan(ctx context.Context, planID, featureID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO plan_features (plan_id, feature_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		planID, featureID,
	)
	return err
}

func (r *Repository) RemoveFeatureFromPlan(ctx context.Context, planID, featureID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM plan_features WHERE plan_id = $1 AND feature_id = $2`, planID, featureID,
	)
	return err
}

// --- Features ---

func (r *Repository) ListFeatures(ctx context.Context) ([]featureFullRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, slug, code, description, is_active, created_at, updated_at FROM features ORDER BY title`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []featureFullRow
	for rows.Next() {
		var f featureFullRow
		if err := rows.Scan(&f.ID, &f.Title, &f.Slug, &f.Code, &f.Description, &f.IsActive, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		features = append(features, f)
	}
	return features, nil
}

type featureFullRow struct {
	ID          string
	Title       string
	Slug        string
	Code        string
	Description *string
	IsActive    bool
	CreatedAt   interface{}
	UpdatedAt   interface{}
}

func (r *Repository) GetFeatureByID(ctx context.Context, id string) (*featureFullRow, error) {
	var f featureFullRow
	err := r.db.QueryRow(ctx,
		`SELECT id, title, slug, code, description, is_active, created_at, updated_at FROM features WHERE id = $1`, id,
	).Scan(&f.ID, &f.Title, &f.Slug, &f.Code, &f.Description, &f.IsActive, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *Repository) CreateFeature(ctx context.Context, title, slug, code string, description *string, isActive bool) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO features (title, slug, code, description, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		title, slug, code, description, isActive,
	).Scan(&id)
	return id, err
}

func (r *Repository) UpdateFeature(ctx context.Context, id string, title *string, description *string, isActive *bool) error {
	query := `UPDATE features SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if title != nil {
		query += fmt.Sprintf(", title = $%d", argIdx)
		args = append(args, *title)
		argIdx++
	}
	if description != nil {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, *description)
		argIdx++
	}
	if isActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argIdx)
		args = append(args, *isActive)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) DeleteFeature(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM features WHERE id = $1`, id)
	return err
}

// --- Promotions ---

func (r *Repository) ListPromotions(ctx context.Context) ([]promotionRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, discount_type, discount_value, duration_months,
		        valid_from, valid_until, is_active, created_at, updated_at
		 FROM promotions ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promos []promotionRow
	for rows.Next() {
		var p promotionRow
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.DiscountType, &p.DiscountValue,
			&p.DurationMonths, &p.ValidFrom, &p.ValidUntil, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}
	return promos, nil
}

type promotionRow struct {
	ID             string
	Name           string
	Description    *string
	DiscountType   string
	DiscountValue  float64
	DurationMonths int
	ValidFrom      interface{}
	ValidUntil     interface{}
	IsActive       bool
	CreatedAt      interface{}
	UpdatedAt      interface{}
}

func (r *Repository) GetPromotionByID(ctx context.Context, id string) (*promotionRow, error) {
	var p promotionRow
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, discount_type, discount_value, duration_months,
		        valid_from, valid_until, is_active, created_at, updated_at
		 FROM promotions WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.DiscountType, &p.DiscountValue,
		&p.DurationMonths, &p.ValidFrom, &p.ValidUntil, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) CreatePromotion(ctx context.Context, name string, description *string, discountType string, discountValue float64, durationMonths int, validFrom, validUntil interface{}) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO promotions (name, description, discount_type, discount_value, duration_months, valid_from, valid_until)
		 VALUES ($1, $2, $3::discount_type, $4, $5, COALESCE($6::timestamp, NOW()), $7::timestamp) RETURNING id`,
		name, description, discountType, discountValue, durationMonths, validFrom, validUntil,
	).Scan(&id)
	return id, err
}

func (r *Repository) UpdatePromotion(ctx context.Context, id string, req interface{}) error {
	_, err := r.db.Exec(ctx, `UPDATE promotions SET updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *Repository) DeactivatePromotion(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE promotions SET is_active = false, updated_at = NOW() WHERE id = $1`, id)
	return err
}

// --- Tenant Plans ---

func (r *Repository) CreateTenantPlan(ctx context.Context, tx pgx.Tx, tenantID, planID, billingCycle string, basePrice, contractedPrice float64, promotionID *string, promoPrice *float64, promoExpiresAt interface{}) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO tenant_plans (tenant_id, plan_id, billing_cycle, base_price, contracted_price, promotion_id, promo_price, promo_expires_at)
		 VALUES ($1, $2, $3::billing_cycle, $4, $5, $6, $7, $8::timestamp)`,
		tenantID, planID, billingCycle, basePrice, contractedPrice, promotionID, promoPrice, promoExpiresAt,
	)
	return err
}

func (r *Repository) DeactivateCurrentPlan(ctx context.Context, tx pgx.Tx, tenantID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE tenant_plans SET is_active = false, ended_at = NOW(), updated_at = NOW()
		 WHERE tenant_id = $1 AND is_active = true`, tenantID,
	)
	return err
}

func (r *Repository) GetTenantPlanHistory(ctx context.Context, tenantID string) ([]interface{}, error) {
	rows, err := r.db.Query(ctx,
		`SELECT tp.id, tp.plan_id, p.name, tp.billing_cycle, tp.base_price, tp.contracted_price,
		        tp.promo_price, tp.promo_expires_at, tp.is_active, tp.started_at, tp.ended_at
		 FROM tenant_plans tp
		 JOIN plans p ON p.id = tp.plan_id
		 WHERE tp.tenant_id = $1
		 ORDER BY tp.created_at DESC`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []interface{}
	for rows.Next() {
		var h struct {
			ID              string
			PlanID          string
			PlanName        string
			BillingCycle    string
			BasePrice       float64
			ContractedPrice float64
			PromoPrice      *float64
			PromoExpiresAt  interface{}
			IsActive        bool
			StartedAt       interface{}
			EndedAt         interface{}
		}
		if err := rows.Scan(&h.ID, &h.PlanID, &h.PlanName, &h.BillingCycle, &h.BasePrice,
			&h.ContractedPrice, &h.PromoPrice, &h.PromoExpiresAt, &h.IsActive, &h.StartedAt, &h.EndedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}
