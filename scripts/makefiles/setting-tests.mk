.PHONY: test-settings-list test-settings-get test-settings-update \
        test-layout-get test-layout-update

# Test settings list
test-settings-list:
	@echo "Testing settings list..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test setting get
test-settings-get:
	@echo "Testing setting get..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings/theme_color \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test setting update
test-settings-update:
	@echo "Testing setting update..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings/theme_color \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"value":"#FF5733"}'
	@echo ""

# Test layout settings GET (returns defaults if none saved)
test-layout-get:
	@echo "Testing layout settings GET..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings/layout \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test layout settings UPDATE
test-layout-update:
	@echo "Testing layout settings UPDATE..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Update layout settings..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings/layout \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"primary_color":"#FF5733","secondary_color":"#33FF57","logo":"https://example.com/logo.png","theme":"Lara"}'; \
	echo ""; echo "2. Read back layout settings..."; \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings/layout \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""
