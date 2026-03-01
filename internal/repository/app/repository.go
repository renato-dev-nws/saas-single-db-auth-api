package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// --- Auth ---

func (r *Repository) GetAppUserByEmail(ctx context.Context, tenantID, email string) (*AppUserRow, error) {
	var u AppUserRow
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, hash_pass, status
		 FROM tenant_app_users WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL`,
		tenantID, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.HashPass, &u.Status)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetAppUserByID(ctx context.Context, tenantID, userID string) (*AppUserRow, error) {
	var u AppUserRow
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, hash_pass, status
		 FROM tenant_app_users WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`,
		tenantID, userID,
	).Scan(&u.ID, &u.Name, &u.Email, &u.HashPass, &u.Status)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type AppUserRow struct {
	ID       string
	Name     string
	Email    string
	HashPass string
	Status   string
}

func (r *Repository) CreateAppUser(ctx context.Context, tenantID, name, email, hashPass string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO tenant_app_users (tenant_id, name, email, hash_pass) VALUES ($1, $2, $3, $4) RETURNING id`,
		tenantID, name, email, hashPass,
	).Scan(&id)
	return id, err
}

func (r *Repository) CreateAppUserProfile(ctx context.Context, appUserID, fullName string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO tenant_app_user_profiles (app_user_id, full_name) VALUES ($1, $2)`,
		appUserID, fullName,
	)
	return err
}

func (r *Repository) GetAppUserProfile(ctx context.Context, appUserID string) (*AppUserProfileRow, error) {
	var p AppUserProfileRow
	err := r.db.QueryRow(ctx,
		`SELECT app_user_id, full_name, phone, document, avatar_url, address, notes
		 FROM tenant_app_user_profiles WHERE app_user_id = $1`,
		appUserID,
	).Scan(&p.AppUserID, &p.FullName, &p.Phone, &p.Document, &p.AvatarURL, &p.Address, &p.Notes)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type AppUserProfileRow struct {
	AppUserID string
	FullName  *string
	Phone     *string
	Document  *string
	AvatarURL *string
	Address   *string
	Notes     *string
}

func (r *Repository) UpdateAppUserProfile(ctx context.Context, appUserID string, fullName, phone, document, address, notes *string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO tenant_app_user_profiles (app_user_id, full_name, phone, document, address, notes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (app_user_id) DO UPDATE SET
		   full_name = COALESCE($2, tenant_app_user_profiles.full_name),
		   phone = COALESCE($3, tenant_app_user_profiles.phone),
		   document = COALESCE($4, tenant_app_user_profiles.document),
		   address = COALESCE($5, tenant_app_user_profiles.address),
		   notes = COALESCE($6, tenant_app_user_profiles.notes),
		   updated_at = NOW()`,
		appUserID, fullName, phone, document, address, notes,
	)
	return err
}

func (r *Repository) UpdateAppUserProfileAvatar(ctx context.Context, appUserID, avatarURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_app_user_profiles SET avatar_url = $1, updated_at = NOW() WHERE app_user_id = $2`,
		avatarURL, appUserID,
	)
	return err
}

func (r *Repository) UpdateAppUserPassword(ctx context.Context, tenantID, userID, hashPass string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tenant_app_users SET hash_pass = $1, updated_at = NOW()
		 WHERE tenant_id = $2 AND id = $3 AND deleted_at IS NULL`,
		hashPass, tenantID, userID,
	)
	return err
}

// --- Catalog (Public) ---

func (r *Repository) ListActiveProducts(ctx context.Context, tenantID string, limit, offset int) ([]interface{}, int64, error) {
	var total int64
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM products WHERE tenant_id = $1 AND is_active = true`, tenantID,
	).Scan(&total)

	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, price, sku, stock, image_url
		 FROM products WHERE tenant_id = $1 AND is_active = true
		 ORDER BY name LIMIT $2 OFFSET $3`, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []interface{}
	for rows.Next() {
		var p struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
			Price       float64 `json:"price"`
			SKU         *string `json:"sku"`
			Stock       int     `json:"stock"`
			ImageURL    *string `json:"image_url"`
		}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.Stock, &p.ImageURL); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, nil
}

func (r *Repository) GetActiveProduct(ctx context.Context, tenantID, productID string) (interface{}, error) {
	var p struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Price       float64 `json:"price"`
		SKU         *string `json:"sku"`
		Stock       int     `json:"stock"`
		ImageURL    *string `json:"image_url"`
	}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, price, sku, stock, image_url
		 FROM products WHERE tenant_id = $1 AND id = $2 AND is_active = true`,
		tenantID, productID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.Stock, &p.ImageURL)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Repository) ListActiveServices(ctx context.Context, tenantID string, limit, offset int) ([]interface{}, int64, error) {
	var total int64
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM services WHERE tenant_id = $1 AND is_active = true`, tenantID,
	).Scan(&total)

	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, price, duration, image_url
		 FROM services WHERE tenant_id = $1 AND is_active = true
		 ORDER BY name LIMIT $2 OFFSET $3`, tenantID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var services []interface{}
	for rows.Next() {
		var s struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
			Price       float64 `json:"price"`
			Duration    *int    `json:"duration"`
			ImageURL    *string `json:"image_url"`
		}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Price, &s.Duration, &s.ImageURL); err != nil {
			return nil, 0, err
		}
		services = append(services, s)
	}
	return services, total, nil
}

func (r *Repository) GetActiveService(ctx context.Context, tenantID, serviceID string) (interface{}, error) {
	var s struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Price       float64 `json:"price"`
		Duration    *int    `json:"duration"`
		ImageURL    *string `json:"image_url"`
	}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, price, duration, image_url
		 FROM services WHERE tenant_id = $1 AND id = $2 AND is_active = true`,
		tenantID, serviceID,
	).Scan(&s.ID, &s.Name, &s.Description, &s.Price, &s.Duration, &s.ImageURL)
	if err != nil {
		return nil, err
	}
	return s, nil
}
