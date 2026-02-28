-- ============================================================
-- SaaS Multi-Tenant (Single Database) — Initial Schema
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Enums
CREATE TYPE tenant_status    AS ENUM ('active', 'suspended', 'cancelled');
CREATE TYPE billing_cycle    AS ENUM ('monthly', 'quarterly', 'semiannual', 'annual');
CREATE TYPE user_status      AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE storage_provider AS ENUM ('local', 's3', 'r2');
CREATE TYPE discount_type    AS ENUM ('percent', 'fixed');
CREATE TYPE plan_type        AS ENUM ('individual', 'business');

-- ============================================================
-- System Admin Tables
-- ============================================================

CREATE TABLE system_admin_users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255),
    email         VARCHAR(255) UNIQUE NOT NULL,
    hash_pass     VARCHAR(255) NOT NULL,
    status        user_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP
);

CREATE TABLE system_admin_profiles (
    admin_user_id UUID         PRIMARY KEY REFERENCES system_admin_users(id) ON DELETE CASCADE,
    full_name     VARCHAR(255),
    title         VARCHAR(255),
    bio           TEXT,
    avatar_url    TEXT,
    social_links  JSONB        NOT NULL DEFAULT '{}',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE system_admin_roles (
    id          SERIAL       PRIMARY KEY,
    title       VARCHAR(100) NOT NULL,
    slug        VARCHAR(50)  UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE system_admin_permissions (
    id          SERIAL       PRIMARY KEY,
    title       VARCHAR(100) NOT NULL,
    slug        VARCHAR(50)  UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE system_admin_user_roles (
    admin_user_id UUID    REFERENCES system_admin_users(id) ON DELETE CASCADE,
    admin_role_id INTEGER REFERENCES system_admin_roles(id) ON DELETE CASCADE,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (admin_user_id, admin_role_id)
);

CREATE TABLE system_admin_role_permissions (
    admin_role_id       INTEGER REFERENCES system_admin_roles(id) ON DELETE CASCADE,
    admin_permission_id INTEGER REFERENCES system_admin_permissions(id) ON DELETE CASCADE,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (admin_role_id, admin_permission_id)
);

-- ============================================================
-- Plans & Features Tables
-- ============================================================

CREATE TABLE features (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    title       VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    code        VARCHAR(10)  UNIQUE NOT NULL,
    description TEXT,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE plans (
    id            UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255)   NOT NULL,
    description   TEXT,
    plan_type     plan_type      NOT NULL DEFAULT 'individual',
    price         DECIMAL(10,2)  NOT NULL DEFAULT 0,
    max_users     INTEGER        NOT NULL DEFAULT 1,
    is_multilang  BOOLEAN        NOT NULL DEFAULT false,
    is_active     BOOLEAN        NOT NULL DEFAULT true,
    created_at    TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE TABLE plan_features (
    plan_id    UUID REFERENCES plans(id)    ON DELETE CASCADE,
    feature_id UUID REFERENCES features(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (plan_id, feature_id)
);

-- ============================================================
-- Tenant Tables
-- ============================================================

CREATE TABLE promotions (
    id              UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255)   NOT NULL,
    description     TEXT,
    discount_type   discount_type  NOT NULL,
    discount_value  DECIMAL(10,2)  NOT NULL,
    duration_months INTEGER        NOT NULL DEFAULT 1,
    valid_from      TIMESTAMP      NOT NULL DEFAULT NOW(),
    valid_until     TIMESTAMP,
    is_active       BOOLEAN        NOT NULL DEFAULT true,
    created_at      TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE TABLE tenants (
    id            UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255)   NOT NULL,
    url_code      CHAR(11)       UNIQUE NOT NULL,
    subdomain     VARCHAR(50)    UNIQUE NOT NULL,
    is_company    BOOLEAN        NOT NULL DEFAULT false,
    company_name  VARCHAR(255),
    custom_domain VARCHAR(255),
    status        tenant_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP      NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP
);

CREATE TABLE tenant_plans (
    id                UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id         UUID           NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id           UUID           NOT NULL REFERENCES plans(id),
    billing_cycle     billing_cycle  NOT NULL DEFAULT 'monthly',
    base_price        DECIMAL(10,2)  NOT NULL,
    contracted_price  DECIMAL(10,2)  NOT NULL,
    price_updated_at  TIMESTAMP      NOT NULL DEFAULT NOW(),
    promotion_id      UUID           REFERENCES promotions(id) ON DELETE SET NULL,
    promo_price       DECIMAL(10,2),
    promo_expires_at  TIMESTAMP,
    is_active         BOOLEAN        NOT NULL DEFAULT true,
    started_at        TIMESTAMP      NOT NULL DEFAULT NOW(),
    ended_at          TIMESTAMP,
    created_at        TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_tenant_plans_active ON tenant_plans(tenant_id) WHERE is_active = true;

CREATE TABLE tenant_profiles (
    tenant_id       UUID         PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    about           TEXT,
    logo_url        TEXT,
    custom_settings JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Tenant Users (Backoffice) Tables
-- ============================================================

CREATE TABLE users (
    id                   UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                 VARCHAR(255) NOT NULL,
    email                VARCHAR(255) UNIQUE NOT NULL,
    hash_pass            VARCHAR(255) NOT NULL,
    last_tenant_url_code CHAR(11),
    email_verified_at    TIMESTAMP,
    status               user_status  NOT NULL DEFAULT 'active',
    created_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMP
);

CREATE TABLE user_profiles (
    user_id    UUID         PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    full_name  VARCHAR(255),
    about      TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id  UUID         REFERENCES tenants(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, slug)
);

CREATE TABLE user_permissions (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    feature_id  UUID         REFERENCES features(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE user_role_permissions (
    role_id       UUID REFERENCES user_roles(id)       ON DELETE CASCADE,
    permission_id UUID REFERENCES user_permissions(id)  ON DELETE CASCADE,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE tenant_members (
    user_id    UUID REFERENCES users(id)      ON DELETE CASCADE,
    tenant_id  UUID REFERENCES tenants(id)    ON DELETE CASCADE,
    role_id    UUID REFERENCES user_roles(id) ON DELETE SET NULL,
    is_owner   BOOLEAN   NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    PRIMARY KEY (user_id, tenant_id)
);

-- ============================================================
-- Tenant Data Tables (Logical Isolation)
-- ============================================================

CREATE TABLE products (
    id          UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID           NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(255)   NOT NULL,
    description TEXT,
    price       DECIMAL(10,2)  NOT NULL DEFAULT 0,
    sku         VARCHAR(100),
    stock       INTEGER        NOT NULL DEFAULT 0,
    is_active   BOOLEAN        NOT NULL DEFAULT true,
    image_url   TEXT,
    created_at  TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP      NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, sku)
);

CREATE TABLE services (
    id          UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID           NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(255)   NOT NULL,
    description TEXT,
    price       DECIMAL(10,2)  NOT NULL DEFAULT 0,
    duration    INTEGER,
    is_active   BOOLEAN        NOT NULL DEFAULT true,
    image_url   TEXT,
    created_at  TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE TABLE settings (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id  UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    category   VARCHAR(100) NOT NULL,
    data       JSONB        NOT NULL DEFAULT '{}',
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, category)
);

-- ============================================================
-- App Users Tables
-- ============================================================

CREATE TABLE tenant_app_users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    hash_pass     VARCHAR(255) NOT NULL,
    status        user_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP,
    UNIQUE (tenant_id, email)
);

CREATE TABLE tenant_app_user_profiles (
    app_user_id UUID         PRIMARY KEY REFERENCES tenant_app_users(id) ON DELETE CASCADE,
    full_name   VARCHAR(255),
    phone       VARCHAR(30),
    document    VARCHAR(30),
    birth_date  DATE,
    avatar_url  TEXT,
    address     JSONB        NOT NULL DEFAULT '{}',
    metadata    JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Email Verification Tokens
-- ============================================================

CREATE TABLE email_verification_tokens (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(64)  UNIQUE NOT NULL,
    expires_at TIMESTAMP    NOT NULL,
    used_at    TIMESTAMP,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Email Templates
-- ============================================================

CREATE TABLE email_templates (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug       VARCHAR(100) UNIQUE NOT NULL,
    subject    VARCHAR(500) NOT NULL,
    body_html  TEXT         NOT NULL,
    variables  JSONB        NOT NULL DEFAULT '[]',
    is_active  BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Images Table (Polymorphic)
-- ============================================================

CREATE TYPE media_type     AS ENUM ('image');
CREATE TYPE image_variant  AS ENUM ('original', 'medium', 'small', 'thumb');

CREATE TABLE images (
    id                 UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id          UUID           NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    imageable_type     VARCHAR(50)    NOT NULL,
    imageable_id       UUID           NOT NULL,
    filename           VARCHAR(255)   NOT NULL,
    original_filename  VARCHAR(255),
    title              VARCHAR(255),
    alt_text           VARCHAR(255),
    media_type         media_type     NOT NULL DEFAULT 'image',
    mime_type          VARCHAR(100)   NOT NULL,
    extension          VARCHAR(10)    NOT NULL,
    variant            image_variant  NOT NULL DEFAULT 'original',
    parent_id          UUID           REFERENCES images(id) ON DELETE CASCADE,
    width              INTEGER,
    height             INTEGER,
    file_size          BIGINT,
    storage_driver     VARCHAR(20)    NOT NULL DEFAULT 'local',
    storage_path       TEXT           NOT NULL,
    public_url         TEXT,
    processing_status  VARCHAR(20)    DEFAULT 'pending',
    processed_at       TIMESTAMP,
    display_order      INTEGER        DEFAULT 0,
    created_at         TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP      NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Indexes
-- ============================================================

CREATE INDEX idx_tenants_url_code    ON tenants(url_code);
CREATE INDEX idx_tenants_subdomain   ON tenants(subdomain);
CREATE INDEX idx_tenants_status      ON tenants(status);
CREATE INDEX idx_tenants_deleted_at  ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX idx_tenant_plans_tenant_id ON tenant_plans(tenant_id);
CREATE INDEX idx_tenant_plans_plan_id   ON tenant_plans(plan_id);

CREATE INDEX idx_promotions_is_active ON promotions(is_active);

CREATE INDEX idx_users_email              ON users(email);
CREATE INDEX idx_users_status             ON users(status);
CREATE INDEX idx_users_deleted_at         ON users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_users_email_verified_at  ON users(email_verified_at) WHERE email_verified_at IS NULL;

CREATE INDEX idx_tenant_members_user_id    ON tenant_members(user_id)   WHERE deleted_at IS NULL;
CREATE INDEX idx_tenant_members_tenant_id  ON tenant_members(tenant_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_products_tenant_id  ON products(tenant_id);
CREATE INDEX idx_services_tenant_id  ON services(tenant_id);
CREATE INDEX idx_settings_tenant_id  ON settings(tenant_id);
CREATE INDEX idx_images_tenant_id         ON images(tenant_id);
CREATE INDEX idx_images_imageable         ON images(imageable_type, imageable_id);
CREATE INDEX idx_images_parent_id         ON images(parent_id);
CREATE INDEX idx_images_processing_status ON images(processing_status) WHERE processing_status IN ('pending', 'processing');

CREATE INDEX idx_tenant_app_users_tenant_id ON tenant_app_users(tenant_id);
CREATE INDEX idx_tenant_app_users_email     ON tenant_app_users(tenant_id, email) WHERE deleted_at IS NULL;

CREATE INDEX idx_system_admin_users_email      ON system_admin_users(email);
CREATE INDEX idx_system_admin_users_status     ON system_admin_users(status);
CREATE INDEX idx_system_admin_users_deleted_at ON system_admin_users(deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_token   ON email_verification_tokens(token);

-- ============================================================
-- Seed Data
-- ============================================================

-- System admin roles
INSERT INTO system_admin_roles (title, slug) VALUES
    ('Super Admin', 'super_admin'),
    ('Admin',       'admin'),
    ('Support',     'support'),
    ('Viewer',      'viewer');

-- System admin permissions
INSERT INTO system_admin_permissions (title, slug) VALUES
    ('Manage Tenants',      'manage_tenants'),
    ('View Tenants',        'view_tenants'),
    ('Manage Plans',        'manage_plans'),
    ('Manage Features',     'manage_features'),
    ('Manage Sys Users',    'manage_sys_users'),
    ('View Analytics',      'view_analytics'),
    ('Manage Billing',      'manage_billing');

-- Assign all permissions to super_admin role
INSERT INTO system_admin_role_permissions (admin_role_id, admin_permission_id)
SELECT r.id, p.id FROM system_admin_roles r, system_admin_permissions p WHERE r.slug = 'super_admin';

-- Assign view + manage tenants to admin role
INSERT INTO system_admin_role_permissions (admin_role_id, admin_permission_id)
SELECT r.id, p.id FROM system_admin_roles r, system_admin_permissions p
WHERE r.slug = 'admin' AND p.slug IN ('manage_tenants', 'view_tenants', 'manage_plans', 'manage_features');

-- Assign view permissions to support role
INSERT INTO system_admin_role_permissions (admin_role_id, admin_permission_id)
SELECT r.id, p.id FROM system_admin_roles r, system_admin_permissions p
WHERE r.slug = 'support' AND p.slug IN ('view_tenants');

-- Assign view to viewer role
INSERT INTO system_admin_role_permissions (admin_role_id, admin_permission_id)
SELECT r.id, p.id FROM system_admin_roles r, system_admin_permissions p
WHERE r.slug = 'viewer' AND p.slug IN ('view_tenants', 'view_analytics');

-- Default admin user (admin@saas.com / admin123)
-- bcrypt hash of 'admin123' with cost 12
INSERT INTO system_admin_users (name, email, hash_pass) VALUES
    ('System Administrator', 'admin@saas.com', '$2a$12$ns1YP4G3P8iRUKwREqMK8eGgIcxvPyAzXxmNibXydt5GRD6LslLG.');

INSERT INTO system_admin_profiles (admin_user_id, full_name, title)
    SELECT id, 'System Administrator', 'Platform Admin' FROM system_admin_users WHERE email = 'admin@saas.com';

-- Assign super_admin role to default admin
INSERT INTO system_admin_user_roles (admin_user_id, admin_role_id)
    SELECT u.id, r.id FROM system_admin_users u, system_admin_roles r
    WHERE u.email = 'admin@saas.com' AND r.slug = 'super_admin';

-- Default features
INSERT INTO features (id, title, slug, code) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Products', 'products', 'prod'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Services', 'services', 'serv');

-- Default promotion
INSERT INTO promotions (id, name, description, discount_type, discount_value, duration_months, valid_from, is_active) VALUES
    ('cc000000-0000-0000-0000-000000000001',
     'Lançamento 50% off',
     '50% de desconto nos primeiros 3 meses',
     'percent', 50.00, 3,
     NOW(), true);

-- ============================================================
-- Default Plans (8 plans)
-- Individual: always 1 user (with and without multilang)
-- Business: 1, 5 and 10 users (with and without multilang)
-- ============================================================
INSERT INTO plans (id, name, description, plan_type, price, max_users, is_multilang) VALUES
    -- Individual plans (1 user always)
    ('10000000-0000-0000-0000-000000000001', 'Individual',           'Plano individual sem multilang',            'individual',  29.90,  1, false),
    ('10000000-0000-0000-0000-000000000002', 'Individual Multi',     'Plano individual com multilang',            'individual',  49.90,  1, true),
    -- Business plans (1, 5, 10 users)
    ('20000000-0000-0000-0000-000000000001', 'Business 1',           'Plano empresarial 1 usuário',               'business',    59.90,  1, false),
    ('20000000-0000-0000-0000-000000000002', 'Business 1 Multi',     'Plano empresarial 1 usuário com multilang', 'business',    79.90,  1, true),
    ('20000000-0000-0000-0000-000000000003', 'Business 5',           'Plano empresarial 5 usuários',              'business',    99.90,  5, false),
    ('20000000-0000-0000-0000-000000000004', 'Business 5 Multi',     'Plano empresarial 5 usuários com multilang','business',   129.90,  5, true),
    ('20000000-0000-0000-0000-000000000005', 'Business 10',          'Plano empresarial 10 usuários',             'business',   199.90, 10, false),
    ('20000000-0000-0000-0000-000000000006', 'Business 10 Multi',    'Plano empresarial 10 usuários com multilang','business',  249.90, 10, true);

-- Assign all features to all plans
INSERT INTO plan_features (plan_id, feature_id)
SELECT p.id, f.id FROM plans p, features f;

-- User permissions (backoffice)
INSERT INTO user_permissions (id, title, slug, feature_id) VALUES
    (uuid_generate_v4(), 'Create Product', 'prod_c', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
    (uuid_generate_v4(), 'Read Product',   'prod_r', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
    (uuid_generate_v4(), 'Update Product', 'prod_u', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
    (uuid_generate_v4(), 'Delete Product', 'prod_d', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
    (uuid_generate_v4(), 'Create Service', 'serv_c', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
    (uuid_generate_v4(), 'Read Service',   'serv_r', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
    (uuid_generate_v4(), 'Update Service', 'serv_u', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
    (uuid_generate_v4(), 'Delete Service', 'serv_d', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
    (uuid_generate_v4(), 'Manage Users',   'user_m', NULL),
    (uuid_generate_v4(), 'Manage Settings','setg_m', NULL);

-- Global user roles (tenant_id NULL = templates copied when creating a tenant)
INSERT INTO user_roles (id, tenant_id, title, slug) VALUES
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', NULL, 'Owner',  'owner'),
    ('ffffffff-ffff-ffff-ffff-ffffffffffff', NULL, 'Admin',  'admin'),
    ('eeeeeeee-eeee-eeee-eeee-111111111111', NULL, 'Member', 'member')
ON CONFLICT (id) DO NOTHING;

-- Assign all permissions to owner template role
INSERT INTO user_role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM user_roles r, user_permissions p WHERE r.slug = 'owner' AND r.tenant_id IS NULL;

-- Assign all permissions to admin template role
INSERT INTO user_role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM user_roles r, user_permissions p WHERE r.slug = 'admin' AND r.tenant_id IS NULL;

-- Assign read-only permissions to member template role
INSERT INTO user_role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM user_roles r, user_permissions p
WHERE r.slug = 'member' AND r.tenant_id IS NULL AND p.slug IN ('prod_r', 'serv_r');

-- ============================================================
-- Default Email Templates
-- ============================================================

INSERT INTO email_templates (slug, subject, body_html, variables) VALUES
(
    'welcome_verify_email',
    'Bem-vindo ao {{app_name}}! Confirme seu e-mail',
    '<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin:0;padding:0;background:#f4f4f7;font-family:Arial,Helvetica,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f7;padding:40px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
  <tr><td style="background:#4F46E5;padding:32px 40px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">{{app_name}}</h1>
  </td></tr>
  <tr><td style="padding:40px;">
    <h2 style="color:#333;margin:0 0 16px;">Olá, {{owner_name}}! 👋</h2>
    <p style="color:#555;font-size:16px;line-height:1.6;">
      Sua conta no <strong>{{app_name}}</strong> foi criada com sucesso!
    </p>
    <p style="color:#555;font-size:16px;line-height:1.6;">
      Seu painel de administração: <strong>{{panel_url}}</strong>
    </p>
    <p style="color:#555;font-size:16px;line-height:1.6;">
      Para ativar sua conta e começar a usar, clique no botão abaixo para confirmar seu e-mail:
    </p>
    <table width="100%" cellpadding="0" cellspacing="0" style="margin:32px 0;">
    <tr><td align="center">
      <a href="{{verify_url}}" style="display:inline-block;background:#4F46E5;color:#ffffff;text-decoration:none;padding:14px 32px;border-radius:6px;font-size:16px;font-weight:bold;">
        Confirmar meu e-mail
      </a>
    </td></tr>
    </table>
    <p style="color:#888;font-size:14px;">
      Se o botão não funcionar, copie e cole este link no seu navegador:<br>
      <a href="{{verify_url}}" style="color:#4F46E5;">{{verify_url}}</a>
    </p>
    <hr style="border:none;border-top:1px solid #eee;margin:32px 0;">
    <p style="color:#888;font-size:13px;margin:0;">
      Detalhes da sua conta:<br>
      <strong>Tenant:</strong> {{tenant_name}}<br>
      <strong>Plano:</strong> {{plan_name}}<br>
      <strong>E-mail:</strong> {{owner_email}}
    </p>
  </td></tr>
  <tr><td style="background:#f8f8fa;padding:24px 40px;text-align:center;">
    <p style="color:#999;font-size:12px;margin:0;">
      Este link expira em 24 horas. Se você não criou essa conta, ignore este e-mail.
    </p>
  </td></tr>
</table>
</td></tr>
</table>
</body>
</html>',
    '["app_name", "owner_name", "owner_email", "tenant_name", "plan_name", "panel_url", "verify_url"]'::jsonb
),
(
    'email_verified',
    '{{app_name}} — E-mail confirmado!',
    '<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin:0;padding:0;background:#f4f4f7;font-family:Arial,Helvetica,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f7;padding:40px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
  <tr><td style="background:#4F46E5;padding:32px 40px;text-align:center;">
    <h1 style="color:#ffffff;margin:0;font-size:24px;">{{app_name}}</h1>
  </td></tr>
  <tr><td style="padding:40px;text-align:center;">
    <h2 style="color:#333;margin:0 0 16px;">E-mail confirmado! ✅</h2>
    <p style="color:#555;font-size:16px;line-height:1.6;">
      Parabéns, <strong>{{owner_name}}</strong>! Seu e-mail foi verificado com sucesso.
    </p>
    <p style="color:#555;font-size:16px;line-height:1.6;">
      Agora você pode acessar todos os recursos do seu plano.
    </p>
    <table width="100%" cellpadding="0" cellspacing="0" style="margin:32px 0;">
    <tr><td align="center">
      <a href="{{panel_url}}" style="display:inline-block;background:#4F46E5;color:#ffffff;text-decoration:none;padding:14px 32px;border-radius:6px;font-size:16px;font-weight:bold;">
        Acessar meu painel
      </a>
    </td></tr>
    </table>
  </td></tr>
  <tr><td style="background:#f8f8fa;padding:24px 40px;text-align:center;">
    <p style="color:#999;font-size:12px;margin:0;">© {{app_name}}</p>
  </td></tr>
</table>
</td></tr>
</table>
</body>
</html>',
    '["app_name", "owner_name", "panel_url"]'::jsonb
);
