.PHONY: test-admin-login test-admin-me test-sysusers-list test-sysusers-create \
        test-plans-list test-plans-create test-features-list test-features-create \
        test-tenants-list test-tenants-create test-promotions-list test-promotions-create

# Test admin login
test-admin-login:
	@echo "Testing admin login..."
	@curl -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}'
	@echo ""

# Test admin me
test-admin-me:
	@echo "Testing admin /me..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8081/api/v1/admin/auth/me \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test sys users
test-sysusers-list:
	@echo "Testing sys users list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8081/api/v1/admin/sys-users \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-sysusers-create:
	@echo "Testing sys user create..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/admin/sys-users \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Admin Teste","email":"admin2@saas.com","password":"admin123"}'
	@echo ""

# Test plans
test-plans-list:
	@echo "Testing plans list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8081/api/v1/admin/plans \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-plans-create:
	@echo "Testing plan create..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/admin/plans \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Plano Teste","plan_type":"business","price":99.90,"max_users":10,"is_multilang":false}'
	@echo ""

# Test features
test-features-list:
	@echo "Testing features list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8081/api/v1/admin/features \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-features-create:
	@echo "Testing feature create..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/admin/features \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Feature Teste","slug":"feature-teste","description":"Teste"}'
	@echo ""

# Test tenants
test-tenants-list:
	@echo "Testing tenants list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8081/api/v1/admin/tenants \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-tenants-create:
	@echo "Testing tenant create via admin..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/admin/tenants \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Test Company","subdomain":"testco","plan_id":"20000000-0000-0000-0000-000000000001","billing_cycle":"monthly","owner_email":"owner@testco.com","owner_full_name":"Test Owner","owner_password":"pass12345"}'
	@echo ""

# Test promotions
test-promotions-list:
	@echo "Testing promotions list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8081/api/v1/admin/promotions \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

test-promotions-create:
	@echo "Testing promotion create..."
	@TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8081/api/v1/admin/promotions \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Promo Teste","slug":"promo-teste","discount_type":"percent","discount_value":10,"duration_months":3,"is_active":true}'
	@echo ""
