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
	@curl -s -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"10000000-0000-0000-0000-000000000001","billing_cycle":"monthly","subdomain":"minhaloja2","is_company":false,"name":"Joao Silva 2","email":"joao2@minha-loja.com","password":"senha12345"}'
	@echo ""

test-subscription-with-promo:
	@echo "Testing subscription with promo..."
	@curl -s -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"20000000-0000-0000-0000-000000000001","billing_cycle":"monthly","promo_code":"Lançamento 50% off","subdomain":"lojapromo2","is_company":true,"company_name":"Loja Promo 2","name":"Maria Promo 2","email":"maria2@loja-promo.com","password":"senha12345"}'
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
		-d '{"plan_id":"20000000-0000-0000-0000-000000000001","billing_cycle":"monthly","subdomain":"novaempresa2","is_company":true,"company_name":"Nova Empresa 2","name":"Novo Usuario 2","email":"novo2@empresa.com","password":"senha12345"}'); \
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
