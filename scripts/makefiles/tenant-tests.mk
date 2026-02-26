.PHONY: test-plans-public test-subscription test-subscription-with-promo \
        test-login test-tenant-login test-tenant test-user-me test-switch-tenant \
        test-tenant-config test-new-tenant test-testenovo

# Test list plans (public)
test-plans-public:
	@curl -X GET http://localhost:8080/api/v1/plans
	@echo ""

# Test subscription (public)
test-subscription:
	@echo "Testing subscription endpoint..."
	@curl -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"10000000-0000-0000-0000-000000000001","billing_cycle":"monthly","tenant_name":"Minha Loja","subdomain":"minhaloja","is_company":false,"owner_name":"Joao Silva","owner_email":"joao@minha-loja.com","owner_password":"senha12345"}'
	@echo ""

test-subscription-with-promo:
	@echo "Testing subscription with promo..."
	@curl -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"20000000-0000-0000-0000-000000000001","billing_cycle":"monthly","promotion_id":"cc000000-0000-0000-0000-000000000001","tenant_name":"Loja Promo","subdomain":"lojapromo","is_company":false,"owner_name":"Maria Promo","owner_email":"maria@loja-promo.com","owner_password":"senha12345"}'
	@echo ""

# Test login backoffice
test-login: test-tenant-login

test-tenant-login:
	@echo "Testing tenant login..."
	@curl -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'
	@echo ""

# Test GET /auth/me
test-tenant: test-user-me

test-user-me:
	@echo "Testing /auth/me..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/auth/me \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test switch tenant (requires test-subscription to have been run first)
test-switch-tenant:
	@echo "Testing switch tenant..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "URL_CODE=$$URL_CODE"; \
	curl -X POST http://localhost:8080/api/v1/auth/switch/$$URL_CODE \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test tenant config (requires test-subscription to have been run first)
test-tenant-config:
	@echo "Testing tenant config..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "URL_CODE=$$URL_CODE"; \
	curl -X GET http://localhost:8080/api/v1/$$URL_CODE/config \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# E2E: create new tenant
test-new-tenant: test-testenovo

test-testenovo:
	@echo "========================================="
	@echo "E2E: Create Tenant + Login + Config"
	@echo "========================================="
	@RESPONSE=$$(curl -s -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"20000000-0000-0000-0000-000000000001","billing_cycle":"monthly","tenant_name":"Nova Empresa","subdomain":"novaempresa","is_company":false,"owner_name":"Novo Usuario","owner_email":"novo@empresa.com","owner_password":"senha12345"}'); \
	echo "1. Subscription:"; echo "$$RESPONSE"; \
	TOKEN=$$(echo "$$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$RESPONSE" | grep -o '"url_code":"[^"]*' | cut -d'"' -f4); \
	echo ""; echo "2. Config:"; \
	curl -X GET http://localhost:8080/api/v1/$$URL_CODE/config \
		-H "Authorization: Bearer $$TOKEN"; \
	echo ""; echo "3. Products:"; \
	curl -X GET http://localhost:8080/api/v1/$$URL_CODE/products \
		-H "Authorization: Bearer $$TOKEN"
	@echo "========================================="
	@echo "Test completed!"
	@echo "========================================="
