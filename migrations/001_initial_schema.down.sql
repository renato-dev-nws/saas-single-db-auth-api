-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS images CASCADE;
DROP TABLE IF EXISTS tenant_app_user_profiles CASCADE;
DROP TABLE IF EXISTS tenant_app_users CASCADE;
DROP TABLE IF EXISTS settings CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS tenant_members CASCADE;
DROP TABLE IF EXISTS user_role_permissions CASCADE;
DROP TABLE IF EXISTS user_permissions CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS user_profiles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS tenant_profiles CASCADE;
DROP TABLE IF EXISTS tenant_plans CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;
DROP TABLE IF EXISTS promotions CASCADE;
DROP TABLE IF EXISTS plan_features CASCADE;
DROP TABLE IF EXISTS plans CASCADE;
DROP TABLE IF EXISTS features CASCADE;
DROP TABLE IF EXISTS system_admin_role_permissions CASCADE;
DROP TABLE IF EXISTS system_admin_user_roles CASCADE;
DROP TABLE IF EXISTS system_admin_permissions CASCADE;
DROP TABLE IF EXISTS system_admin_roles CASCADE;
DROP TABLE IF EXISTS system_admin_profiles CASCADE;
DROP TABLE IF EXISTS system_admin_users CASCADE;

-- Drop enums
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS storage_provider;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS billing_cycle;
DROP TYPE IF EXISTS tenant_status;
