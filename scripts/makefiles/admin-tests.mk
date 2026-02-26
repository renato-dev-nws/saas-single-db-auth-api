ADMIN_URL=http://localhost:8081
ADMIN_EMAIL=admin@saas.com
ADMIN_PASS=admin123

# ─── Helpers ───────────────────────────────────────────────
define get_admin_token
$(shell curl -s -X POST $(ADMIN_URL)/api/v1/admin/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"$(ADMIN_EMAIL)","password":"$(ADMIN_PASS)"}' \
    | grep -o '"token":"[^"]*' | cut -d'"' -f4)
endef

# ─── Auth ──────────────────────────────────────────────────
test-admin-login:
	@curl -s -X POST $(ADMIN_URL)/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"$(ADMIN_EMAIL)","password":"$(ADMIN_PASS)"}' | jq .
	@echo ""

test-admin-me:
	@TOKEN=$(call get_admin_token); \
	curl -s $(ADMIN_URL)/api/v1/admin/auth/me \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

# ─── Sys Users ─────────────────────────────────────────────
test-sysusers-list:
	@TOKEN=$(call get_admin_token); \
	curl -s $(ADMIN_URL)/api/v1/admin/sys-users \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-sysusers-create:
	@TOKEN=$(call get_admin_token); \
	curl -s -X POST $(ADMIN_URL)/api/v1/admin/sys-users \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"email":"support@saas.com","password":"suporte123","full_name":"Suporte SaaS","role_slug":"support"}' | jq .
	@echo ""

# ─── Plans ─────────────────────────────────────────────────
test-plans-list:
	@TOKEN=$(call get_admin_token); \
	curl -s $(ADMIN_URL)/api/v1/admin/plans \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-plans-create:
	@TOKEN=$(call get_admin_token); \
	curl -s -X POST $(ADMIN_URL)/api/v1/admin/plans \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Enterprise","description":"Plano completo","price":199.90,"feature_ids":["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"]}' | jq .
	@echo ""

# ─── Features ──────────────────────────────────────────────
test-features-list:
	@TOKEN=$(call get_admin_token); \
	curl -s $(ADMIN_URL)/api/v1/admin/features \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-features-create:
	@TOKEN=$(call get_admin_token); \
	curl -s -X POST $(ADMIN_URL)/api/v1/admin/features \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"title":"Blog","slug":"blog","code":"blog","description":"Módulo de blog","is_active":true}' | jq .
	@echo ""

# ─── Tenants ───────────────────────────────────────────────
test-tenants-list:
	@TOKEN=$(call get_admin_token); \
	curl -s "$(ADMIN_URL)/api/v1/admin/tenants?page=1&page_size=10" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-tenants-create:
	@TOKEN=$(call get_admin_token); \
	curl -s -X POST $(ADMIN_URL)/api/v1/admin/tenants \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"url_code":"empresa-teste","subdomain":"empresa-teste","plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","company_name":"Empresa Teste Ltda","is_company":true,"owner_email":"dono@empresa-teste.com","owner_full_name":"Dono Empresa","owner_password":"senha12345"}' | jq .
	@echo ""

# ─── Promotions ────────────────────────────────────────────
test-promotions-list:
	@TOKEN=$(call get_admin_token); \
	curl -s $(ADMIN_URL)/api/v1/admin/promotions \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-promotions-create:
	@TOKEN=$(call get_admin_token); \
	curl -s -X POST $(ADMIN_URL)/api/v1/admin/promotions \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Black Friday 30%","description":"30% off por 2 meses","discount_type":"percent","discount_value":30.00,"duration_months":2,"valid_from":"2026-11-01T00:00:00Z","valid_until":"2026-11-30T23:59:59Z"}' | jq .
	@echo ""
