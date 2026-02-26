package tenant

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

// --- Subscription ---

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
	_, err := tx.Exec(ctx, `INSERT INTO tenant_profiles (tenant_id) VALUES ($1)`, tenantID)
	return err
}

func (r *Repository) CreateTenantPlan(ctx context.Context, tx pgx.Tx, tenantID, planID, billingCycle string, basePrice, contractedPrice float64, promotionID *string, promoPrice *float64, promoExpiresAt interface{}) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO tenant_plans (tenant_id, plan_id, billing_cycle, base_price, contracted_price, promotion_id, promo_price, promo_expires_at)
		 VALUES ($1, $2, $3::billing_cycle, $4, $5, $6, $7, $8::timestamp)`,
		tenantID, planID, billingCycle, basePrice, contractedPrice, promotionID, promoPrice, promoExpiresAt,
	)
	return err
}

func (r *Repository) GetPlanByID(ctx context.Context, id string) (*planRow, error) {
	var p planRow
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, price, max_users, is_multilang, is_active FROM plans WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.MaxUsers, &p.IsMultilang, &p.IsActive)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type planRow struct {
	ID          string
	Name        string
	Description *string
	Price       float64
	MaxUsers    int
	IsMultilang bool
	IsActive    bool
}

func (r *Repository) GetPromotionByID(ctx context.Context, id string) (*promoRow, error) {
	var p promoRow
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, discount_type, discount_value, duration_months, valid_from, valid_until, is_active
		 FROM promotions WHERE id = $1 AND is_active = true`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.DiscountType, &p.DiscountValue, &p.DurationMonths, &p.ValidFrom, &p.ValidUntil, &p.IsActive)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type promoRow struct {
	ID             string
	Name           string
	Description    *string
	DiscountType   string
	DiscountValue  float64
	DurationMonths int
	ValidFrom      interface{}
	ValidUntil     interface{}
	IsActive       bool
}

func (r *Repository) ListActivePlans(ctx context.Context) ([]interface{}, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.name, p.description, p.price, p.max_users, p.is_multilang
		 FROM plans p WHERE p.is_active = true ORDER BY p.price`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []interface{}
	for rows.Next() {
		var p struct {
			ID          string   `json:"id"`
			Name        string   `json:"name"`
			Description *string  `json:"description"`
			Price       float64  `json:"price"`
			MaxUsers    int      `json:"max_users"`
			IsMultilang bool     `json:"is_multilang"`
			Features    []string `json:"features"`
		}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.MaxUsers, &p.IsMultilang); err != nil {
			return nil, err
		}
		// Get features
		frows, _ := r.db.Query(ctx,
			`SELECT f.slug FROM features f JOIN plan_features pf ON pf.feature_id = f.id WHERE pf.plan_id = $1 AND f.is_active = true`, p.ID)
		if frows != nil {
			for frows.Next() {
				var slug string
				frows.Scan(&slug)
				p.Features = append(p.Features, slug)
			}
			frows.Close()
		}
		plans = append(plans, p)
	}
	return plans, nil
}

// --- User operations (for subscription) ---

func (r *Repository) CreateUser(ctx context.Context, tx pgx.Tx, name, email, hashPass, urlCode string) (string, error) {
	var id string
	err := tx.QueryRow(ctx,
		`INSERT INTO users (name, email, hash_pass, last_tenant_url_code)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		name, email, hashPass, urlCode,
	).Scan(&id)
	return id, err
}

func (r *Repository) CreateUserProfile(ctx context.Context, tx pgx.Tx, userID, fullName string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO user_profiles (user_id, full_name) VALUES ($1, $2)`, userID, fullName,
	)
	return err
}

func (r *Repository) CopyGlobalRolesToTenant(ctx context.Context, tx pgx.Tx, tenantID string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO user_roles (tenant_id, title, slug)
		 SELECT $1, title, slug FROM user_roles WHERE tenant_id IS NULL`,
		tenantID,
	)
	if err != nil {
		return err
	}

	// Copy role permissions from global templates to new tenant roles
	_, err = tx.Exec(ctx,
		`INSERT INTO user_role_permissions (role_id, permission_id)
		 SELECT tr.id, grp.permission_id
		 FROM user_roles tr
		 JOIN user_roles gr ON gr.slug = tr.slug AND gr.tenant_id IS NULL
		 JOIN user_role_permissions grp ON grp.role_id = gr.id
		 WHERE tr.tenant_id = $1`,
		tenantID,
	)
	return err
}

func (r *Repository) GetTenantRoleBySlug(ctx context.Context, tenantID, slug string) (string, error) {
	var roleID string
	err := r.db.QueryRow(ctx,
		`SELECT id FROM user_roles WHERE tenant_id = $1 AND slug = $2`, tenantID, slug,
	).Scan(&roleID)
	return roleID, err
}

func (r *Repository) GetTenantRoleBySlugTx(ctx context.Context, tx pgx.Tx, tenantID, slug string) (string, error) {
	var roleID string
	err := tx.QueryRow(ctx,
		`SELECT id FROM user_roles WHERE tenant_id = $1 AND slug = $2`, tenantID, slug,
	).Scan(&roleID)
	return roleID, err
}

func (r *Repository) CreateTenantMember(ctx context.Context, tx pgx.Tx, userID, tenantID string, roleID *string, isOwner bool) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO tenant_members (user_id, tenant_id, role_id, is_owner)
		 VALUES ($1, $2, $3, $4)`,
		userID, tenantID, roleID, isOwner,
	)
	return err
}

// --- Auth ---

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*userRow, error) {
	var u userRow
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, hash_pass, last_tenant_url_code, status
		 FROM users WHERE email = $1 AND deleted_at IS NULL`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.HashPass, &u.LastTenantURLCode, &u.Status)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*userRow, error) {
	var u userRow
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, hash_pass, last_tenant_url_code, status
		 FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.HashPass, &u.LastTenantURLCode, &u.Status)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type userRow struct {
	ID                string
	Name              string
	Email             string
	HashPass          string
	LastTenantURLCode *string
	Status            string
}

func (r *Repository) GetUserProfile(ctx context.Context, userID string) (*profileRow, error) {
	var p profileRow
	err := r.db.QueryRow(ctx,
		`SELECT user_id, full_name, about, avatar_url FROM user_profiles WHERE user_id = $1`, userID,
	).Scan(&p.UserID, &p.FullName, &p.About, &p.AvatarURL)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type profileRow struct {
	UserID    string
	FullName  *string
	About     *string
	AvatarURL *string
}

func (r *Repository) UpdateUserProfile(ctx context.Context, userID string, fullName, about *string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_profiles (user_id, full_name, about)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET
		   full_name = COALESCE($2, user_profiles.full_name),
		   about = COALESCE($3, user_profiles.about),
		   updated_at = NOW()`,
		userID, fullName, about,
	)
	return err
}

func (r *Repository) UpdateUserProfileAvatar(ctx context.Context, userID, avatarURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE user_profiles SET avatar_url = $1, updated_at = NOW() WHERE user_id = $2`,
		avatarURL, userID,
	)
	return err
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID, hashPass string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET hash_pass = $1, updated_at = NOW() WHERE id = $2`, hashPass, userID,
	)
	return err
}

func (r *Repository) UpdateUserLastTenant(ctx context.Context, userID, urlCode string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET last_tenant_url_code = $1, updated_at = NOW() WHERE id = $2`, urlCode, userID,
	)
	return err
}

func (r *Repository) GetUserTenants(ctx context.Context, userID string) ([]tenantBrief, error) {
	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.url_code, t.name, tm.is_owner
		 FROM tenant_members tm
		 JOIN tenants t ON t.id = tm.tenant_id
		 WHERE tm.user_id = $1 AND tm.deleted_at IS NULL AND t.deleted_at IS NULL AND t.status = 'active'
		 ORDER BY tm.is_owner DESC, t.name`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []tenantBrief
	for rows.Next() {
		var t tenantBrief
		if err := rows.Scan(&t.ID, &t.URLCode, &t.Name, &t.IsOwner); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	return tenants, nil
}

type tenantBrief struct {
	ID      string `json:"id"`
	URLCode string `json:"url_code"`
	Name    string `json:"name"`
	IsOwner bool   `json:"is_owner"`
}

// TenantBriefExported is the exported version for use by services
type TenantBriefExported struct {
	ID      string `json:"id"`
	URLCode string `json:"url_code"`
	Name    string `json:"name"`
	IsOwner bool   `json:"is_owner"`
}

func (r *Repository) GetTenantByURLCode(ctx context.Context, urlCode string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`SELECT id FROM tenants WHERE url_code = $1 AND deleted_at IS NULL AND status = 'active'`, urlCode,
	).Scan(&id)
	return id, err
}

func (r *Repository) GetTenantByID(ctx context.Context, id string) (*tenantDetail, error) {
	var t tenantDetail
	err := r.db.QueryRow(ctx,
		`SELECT id, name, url_code, subdomain, is_company, company_name, custom_domain, status
		 FROM tenants WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&t.ID, &t.Name, &t.URLCode, &t.Subdomain, &t.IsCompany, &t.CompanyName, &t.CustomDomain, &t.Status)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

type tenantDetail struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	URLCode      string  `json:"url_code"`
	Subdomain    string  `json:"subdomain"`
	IsCompany    bool    `json:"is_company"`
	CompanyName  *string `json:"company_name"`
	CustomDomain *string `json:"custom_domain"`
	Status       string  `json:"status"`
}

func (r *Repository) GetTenantFeatures(ctx context.Context, tenantID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT f.slug
		 FROM tenant_plans tp
		 JOIN plan_features pf ON pf.plan_id = tp.plan_id
		 JOIN features f ON f.id = pf.feature_id
		 WHERE tp.tenant_id = $1 AND tp.is_active = true AND f.is_active = true`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []string
	for rows.Next() {
		var slug string
		rows.Scan(&slug)
		features = append(features, slug)
	}
	return features, nil
}

func (r *Repository) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT p.slug
		 FROM user_permissions p
		 JOIN user_role_permissions rp ON rp.permission_id = p.id
		 JOIN user_roles r ON r.id = rp.role_id
		 JOIN tenant_members tm ON tm.role_id = r.id
		 WHERE tm.user_id = $1 AND tm.tenant_id = $2 AND tm.deleted_at IS NULL
		   AND r.tenant_id = $2`, userID, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var slug string
		rows.Scan(&slug)
		perms = append(perms, slug)
	}
	return perms, nil
}

func (r *Repository) IsMember(ctx context.Context, userID, tenantID string) bool {
	var exists bool
	r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tenant_members WHERE user_id = $1 AND tenant_id = $2 AND deleted_at IS NULL)`,
		userID, tenantID,
	).Scan(&exists)
	return exists
}

// --- Tenant Config ---

func (r *Repository) GetActiveTenantPlan(ctx context.Context, tenantID string) (*activePlanRow, error) {
	var p activePlanRow
	err := r.db.QueryRow(ctx,
		`SELECT tp.id, tp.plan_id, pl.name, pl.max_users, pl.is_multilang,
		        tp.billing_cycle, tp.contracted_price, tp.promo_price, tp.promo_expires_at,
		        tp.price_updated_at
		 FROM tenant_plans tp
		 JOIN plans pl ON pl.id = tp.plan_id
		 WHERE tp.tenant_id = $1 AND tp.is_active = true
		 LIMIT 1`, tenantID,
	).Scan(&p.ID, &p.PlanID, &p.PlanName, &p.MaxUsers, &p.IsMultilang,
		&p.BillingCycle, &p.ContractedPrice, &p.PromoPrice, &p.PromoExpiresAt, &p.PriceUpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type activePlanRow struct {
	ID              string
	PlanID          string
	PlanName        string
	MaxUsers        int
	IsMultilang     bool
	BillingCycle    string
	ContractedPrice float64
	PromoPrice      *float64
	PromoExpiresAt  interface{}
	PriceUpdatedAt  interface{}
}

func (r *Repository) CountTenantMembers(ctx context.Context, tenantID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenant_members WHERE tenant_id = $1 AND deleted_at IS NULL`, tenantID,
	).Scan(&count)
	return count, err
}

// --- Tenant Profile ---

func (r *Repository) GetTenantProfile(ctx context.Context, tenantID string) (*tenantProfileRow, error) {
	var p tenantProfileRow
	err := r.db.QueryRow(ctx,
		`SELECT tenant_id, about, logo_url, custom_settings FROM tenant_profiles WHERE tenant_id = $1`, tenantID,
	).Scan(&p.TenantID, &p.About, &p.LogoURL, &p.CustomSettings)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type tenantProfileRow struct {
	TenantID       string
	About          *string
	LogoURL        *string
	CustomSettings interface{}
}

func (r *Repository) UpdateTenantProfile(ctx context.Context, tenantID string, about *string, customSettings interface{}) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_profiles SET about = COALESCE($2, about), custom_settings = COALESCE($3::jsonb, custom_settings), updated_at = NOW()
		 WHERE tenant_id = $1`, tenantID, about, customSettings,
	)
	return err
}

func (r *Repository) UpdateTenantLogo(ctx context.Context, tenantID, logoURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_profiles SET logo_url = $1, updated_at = NOW() WHERE tenant_id = $2`, logoURL, tenantID,
	)
	return err
}

// --- Members ---

func (r *Repository) ListTenantMembers(ctx context.Context, tenantID string) ([]memberRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT u.id, u.name, u.email, u.status, tm.is_owner,
		        r.id, r.title, r.slug,
		        up.full_name, up.avatar_url, tm.created_at
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
			&m.FullName, &m.AvatarURL, &m.CreatedAt); err != nil {
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
	CreatedAt interface{}
}

func (r *Repository) GetMember(ctx context.Context, tenantID, userID string) (*memberRow, error) {
	var m memberRow
	err := r.db.QueryRow(ctx,
		`SELECT u.id, u.name, u.email, u.status, tm.is_owner,
		        r.id, r.title, r.slug,
		        up.full_name, up.avatar_url, tm.created_at
		 FROM tenant_members tm
		 JOIN users u ON u.id = tm.user_id
		 LEFT JOIN user_roles r ON r.id = tm.role_id
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 WHERE tm.tenant_id = $1 AND tm.user_id = $2 AND tm.deleted_at IS NULL AND u.deleted_at IS NULL`,
		tenantID, userID,
	).Scan(&m.UserID, &m.Name, &m.Email, &m.Status, &m.IsOwner,
		&m.RoleID, &m.RoleTitle, &m.RoleSlug,
		&m.FullName, &m.AvatarURL, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) UpdateMemberRole(ctx context.Context, tenantID, userID, roleID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_members SET role_id = $1, updated_at = NOW()
		 WHERE tenant_id = $2 AND user_id = $3 AND deleted_at IS NULL`, roleID, tenantID, userID,
	)
	return err
}

func (r *Repository) RemoveMember(ctx context.Context, tenantID, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_members SET deleted_at = NOW(), updated_at = NOW()
		 WHERE tenant_id = $1 AND user_id = $2 AND deleted_at IS NULL`, tenantID, userID,
	)
	return err
}

// --- Roles ---

func (r *Repository) ListTenantRoles(ctx context.Context, tenantID string) ([]roleRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, slug FROM user_roles WHERE tenant_id = $1 ORDER BY title`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []roleRow
	for rows.Next() {
		var role roleRow
		if err := rows.Scan(&role.ID, &role.Title, &role.Slug); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

type roleRow struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

func (r *Repository) GetTenantRoleByID(ctx context.Context, tenantID, roleID string) (*roleRow, error) {
	var role roleRow
	err := r.db.QueryRow(ctx,
		`SELECT id, title, slug FROM user_roles WHERE tenant_id = $1 AND id = $2`, tenantID, roleID,
	).Scan(&role.ID, &role.Title, &role.Slug)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) CreateTenantRole(ctx context.Context, tenantID, title, slug string) (*roleRow, error) {
	var role roleRow
	err := r.db.QueryRow(ctx,
		`INSERT INTO user_roles (tenant_id, title, slug) VALUES ($1, $2, $3) RETURNING id, title, slug`,
		tenantID, title, slug,
	).Scan(&role.ID, &role.Title, &role.Slug)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) UpdateTenantRole(ctx context.Context, tenantID, roleID string, title *string) error {
	query := `UPDATE user_roles SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if title != nil {
		query += fmt.Sprintf(", title = $%d", argIdx)
		args = append(args, *title)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE tenant_id = $%d AND id = $%d", argIdx, argIdx+1)
	args = append(args, tenantID, roleID)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) DeleteTenantRole(ctx context.Context, tenantID, roleID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM user_roles WHERE tenant_id = $1 AND id = $2`, tenantID, roleID,
	)
	return err
}

func (r *Repository) GetRolePermissions(ctx context.Context, roleID string) ([]permissionRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.title, p.slug, p.description
		 FROM user_permissions p
		 JOIN user_role_permissions rp ON rp.permission_id = p.id
		 WHERE rp.role_id = $1 ORDER BY p.slug`, roleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []permissionRow
	for rows.Next() {
		var p permissionRow
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Description); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

type permissionRow struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

func (r *Repository) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		roleID, permissionID,
	)
	return err
}

func (r *Repository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM user_role_permissions WHERE role_id = $1 AND permission_id = $2`, roleID, permissionID,
	)
	return err
}

// --- Products ---

func (r *Repository) ListProducts(ctx context.Context, tenantID string, limit, offset int) ([]interface{}, int64, error) {
	var total int64
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM products WHERE tenant_id = $1`, tenantID,
	).Scan(&total)

	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, price, sku, stock, is_active, image_url, created_at, updated_at
		 FROM products WHERE tenant_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []interface{}
	for rows.Next() {
		var p struct {
			ID          string      `json:"id"`
			Name        string      `json:"name"`
			Description *string     `json:"description"`
			Price       float64     `json:"price"`
			SKU         *string     `json:"sku"`
			Stock       int         `json:"stock"`
			IsActive    bool        `json:"is_active"`
			ImageURL    *string     `json:"image_url"`
			CreatedAt   interface{} `json:"created_at"`
			UpdatedAt   interface{} `json:"updated_at"`
		}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.Stock, &p.IsActive, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, nil
}

func (r *Repository) GetProduct(ctx context.Context, tenantID, productID string) (interface{}, error) {
	var p struct {
		ID          string      `json:"id"`
		Name        string      `json:"name"`
		Description *string     `json:"description"`
		Price       float64     `json:"price"`
		SKU         *string     `json:"sku"`
		Stock       int         `json:"stock"`
		IsActive    bool        `json:"is_active"`
		ImageURL    *string     `json:"image_url"`
		CreatedAt   interface{} `json:"created_at"`
		UpdatedAt   interface{} `json:"updated_at"`
	}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, price, sku, stock, is_active, image_url, created_at, updated_at
		 FROM products WHERE tenant_id = $1 AND id = $2`, tenantID, productID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.Stock, &p.IsActive, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Repository) CreateProduct(ctx context.Context, tenantID, name string, description *string, price float64, sku *string, stock int) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO products (tenant_id, name, description, price, sku, stock) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		tenantID, name, description, price, sku, stock,
	).Scan(&id)
	return id, err
}

func (r *Repository) UpdateProduct(ctx context.Context, tenantID, productID string, name *string, description *string, price *float64, sku *string, stock *int, isActive *bool) error {
	query := `UPDATE products SET updated_at = NOW()`
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
	if sku != nil {
		query += fmt.Sprintf(", sku = $%d", argIdx)
		args = append(args, *sku)
		argIdx++
	}
	if stock != nil {
		query += fmt.Sprintf(", stock = $%d", argIdx)
		args = append(args, *stock)
		argIdx++
	}
	if isActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argIdx)
		args = append(args, *isActive)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE tenant_id = $%d AND id = $%d", argIdx, argIdx+1)
	args = append(args, tenantID, productID)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) DeleteProduct(ctx context.Context, tenantID, productID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM products WHERE tenant_id = $1 AND id = $2`, tenantID, productID,
	)
	return err
}

func (r *Repository) UpdateProductImage(ctx context.Context, tenantID, productID, imageURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE products SET image_url = $1, updated_at = NOW() WHERE tenant_id = $2 AND id = $3`,
		imageURL, tenantID, productID,
	)
	return err
}

// --- Services ---

func (r *Repository) ListServices(ctx context.Context, tenantID string, limit, offset int) ([]interface{}, int64, error) {
	var total int64
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM services WHERE tenant_id = $1`, tenantID,
	).Scan(&total)

	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, price, duration, is_active, image_url, created_at, updated_at
		 FROM services WHERE tenant_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var services []interface{}
	for rows.Next() {
		var s struct {
			ID          string      `json:"id"`
			Name        string      `json:"name"`
			Description *string     `json:"description"`
			Price       float64     `json:"price"`
			Duration    *int        `json:"duration"`
			IsActive    bool        `json:"is_active"`
			ImageURL    *string     `json:"image_url"`
			CreatedAt   interface{} `json:"created_at"`
			UpdatedAt   interface{} `json:"updated_at"`
		}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Price, &s.Duration, &s.IsActive, &s.ImageURL, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		services = append(services, s)
	}
	return services, total, nil
}

func (r *Repository) GetService(ctx context.Context, tenantID, serviceID string) (interface{}, error) {
	var s struct {
		ID          string      `json:"id"`
		Name        string      `json:"name"`
		Description *string     `json:"description"`
		Price       float64     `json:"price"`
		Duration    *int        `json:"duration"`
		IsActive    bool        `json:"is_active"`
		ImageURL    *string     `json:"image_url"`
		CreatedAt   interface{} `json:"created_at"`
		UpdatedAt   interface{} `json:"updated_at"`
	}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, price, duration, is_active, image_url, created_at, updated_at
		 FROM services WHERE tenant_id = $1 AND id = $2`, tenantID, serviceID,
	).Scan(&s.ID, &s.Name, &s.Description, &s.Price, &s.Duration, &s.IsActive, &s.ImageURL, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) CreateService(ctx context.Context, tenantID, name string, description *string, price float64, duration *int) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO services (tenant_id, name, description, price, duration) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		tenantID, name, description, price, duration,
	).Scan(&id)
	return id, err
}

func (r *Repository) UpdateService(ctx context.Context, tenantID, serviceID string, name *string, description *string, price *float64, duration *int, isActive *bool) error {
	query := `UPDATE services SET updated_at = NOW()`
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
	if duration != nil {
		query += fmt.Sprintf(", duration = $%d", argIdx)
		args = append(args, *duration)
		argIdx++
	}
	if isActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argIdx)
		args = append(args, *isActive)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE tenant_id = $%d AND id = $%d", argIdx, argIdx+1)
	args = append(args, tenantID, serviceID)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) DeleteService(ctx context.Context, tenantID, serviceID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM services WHERE tenant_id = $1 AND id = $2`, tenantID, serviceID,
	)
	return err
}

func (r *Repository) UpdateServiceImage(ctx context.Context, tenantID, serviceID, imageURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE services SET image_url = $1, updated_at = NOW() WHERE tenant_id = $2 AND id = $3`,
		imageURL, tenantID, serviceID,
	)
	return err
}

// --- Settings ---

func (r *Repository) ListSettings(ctx context.Context, tenantID string) ([]settingRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, category, data, created_at, updated_at
		 FROM settings WHERE tenant_id = $1 ORDER BY category`, tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []settingRow
	for rows.Next() {
		var s settingRow
		if err := rows.Scan(&s.ID, &s.Category, &s.Data, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

type settingRow struct {
	ID        string      `json:"id"`
	Category  string      `json:"category"`
	Data      interface{} `json:"data"`
	CreatedAt interface{} `json:"created_at"`
	UpdatedAt interface{} `json:"updated_at"`
}

func (r *Repository) GetSetting(ctx context.Context, tenantID, category string) (*settingRow, error) {
	var s settingRow
	err := r.db.QueryRow(ctx,
		`SELECT id, category, data, created_at, updated_at
		 FROM settings WHERE tenant_id = $1 AND category = $2`, tenantID, category,
	).Scan(&s.ID, &s.Category, &s.Data, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) UpsertSetting(ctx context.Context, tenantID, category string, data interface{}) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO settings (tenant_id, category, data) VALUES ($1, $2, $3::jsonb)
		 ON CONFLICT (tenant_id, category) DO UPDATE SET data = $3::jsonb, updated_at = NOW()`,
		tenantID, category, data,
	)
	return err
}

// --- Images ---

func (r *Repository) CreateImage(ctx context.Context, tenantID, originalName, storagePath, publicURL string, fileSize int64, mimeType, provider string, entityType *string, entityID *string, uploadedBy *string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO images (tenant_id, original_name, storage_path, public_url, file_size, mime_type, provider, entity_type, entity_id, uploaded_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::storage_provider, $8, $9, $10) RETURNING id`,
		tenantID, originalName, storagePath, publicURL, fileSize, mimeType, provider, entityType, entityID, uploadedBy,
	).Scan(&id)
	return id, err
}

func (r *Repository) ListImages(ctx context.Context, tenantID string, limit, offset int) ([]interface{}, int64, error) {
	var total int64
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM images WHERE tenant_id = $1`, tenantID).Scan(&total)

	rows, err := r.db.Query(ctx,
		`SELECT id, original_name, storage_path, public_url, file_size, mime_type, entity_type, entity_id, created_at
		 FROM images WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var images []interface{}
	for rows.Next() {
		var i struct {
			ID           string      `json:"id"`
			OriginalName string      `json:"original_name"`
			StoragePath  string      `json:"storage_path"`
			PublicURL    string      `json:"public_url"`
			FileSize     int64       `json:"file_size"`
			MimeType     string      `json:"mime_type"`
			EntityType   *string     `json:"entity_type"`
			EntityID     *string     `json:"entity_id"`
			CreatedAt    interface{} `json:"created_at"`
		}
		if err := rows.Scan(&i.ID, &i.OriginalName, &i.StoragePath, &i.PublicURL, &i.FileSize, &i.MimeType, &i.EntityType, &i.EntityID, &i.CreatedAt); err != nil {
			return nil, 0, err
		}
		images = append(images, i)
	}
	return images, total, nil
}

func (r *Repository) DeleteImage(ctx context.Context, tenantID, imageID string) (string, error) {
	var storagePath string
	err := r.db.QueryRow(ctx,
		`DELETE FROM images WHERE tenant_id = $1 AND id = $2 RETURNING storage_path`, tenantID, imageID,
	).Scan(&storagePath)
	return storagePath, err
}

// --- App Users (managed by backoffice) ---

func (r *Repository) ListAppUsers(ctx context.Context, tenantID string, limit, offset int) ([]interface{}, int64, error) {
	var total int64
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenant_app_users WHERE tenant_id = $1 AND deleted_at IS NULL`, tenantID,
	).Scan(&total)

	rows, err := r.db.Query(ctx,
		`SELECT u.id, u.name, u.email, u.status, u.created_at,
		        p.full_name, p.phone
		 FROM tenant_app_users u
		 LEFT JOIN tenant_app_user_profiles p ON p.app_user_id = u.id
		 WHERE u.tenant_id = $1 AND u.deleted_at IS NULL
		 ORDER BY u.created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []interface{}
	for rows.Next() {
		var u struct {
			ID        string      `json:"id"`
			Name      string      `json:"name"`
			Email     string      `json:"email"`
			Status    string      `json:"status"`
			CreatedAt interface{} `json:"created_at"`
			FullName  *string     `json:"full_name"`
			Phone     *string     `json:"phone"`
		}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Status, &u.CreatedAt, &u.FullName, &u.Phone); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *Repository) GetAppUser(ctx context.Context, tenantID, userID string) (interface{}, error) {
	var u struct {
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		Email     string      `json:"email"`
		Status    string      `json:"status"`
		CreatedAt interface{} `json:"created_at"`
		FullName  *string     `json:"full_name"`
		Phone     *string     `json:"phone"`
		Document  *string     `json:"document"`
	}
	err := r.db.QueryRow(ctx,
		`SELECT u.id, u.name, u.email, u.status, u.created_at,
		        p.full_name, p.phone, p.document
		 FROM tenant_app_users u
		 LEFT JOIN tenant_app_user_profiles p ON p.app_user_id = u.id
		 WHERE u.tenant_id = $1 AND u.id = $2 AND u.deleted_at IS NULL`, tenantID, userID,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Status, &u.CreatedAt, &u.FullName, &u.Phone, &u.Document)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) UpdateAppUserStatus(ctx context.Context, tenantID, userID, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_app_users SET status = $1, updated_at = NOW()
		 WHERE tenant_id = $2 AND id = $3 AND deleted_at IS NULL`, status, tenantID, userID,
	)
	return err
}

func (r *Repository) SoftDeleteAppUser(ctx context.Context, tenantID, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_app_users SET deleted_at = NOW(), updated_at = NOW()
		 WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`, tenantID, userID,
	)
	return err
}
