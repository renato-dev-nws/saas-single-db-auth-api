.PHONY: test-settings-get test-settings-update \
        test-layout-get test-layout-update

# Test settings GET (returns layout + convert_webp)
test-settings-get:
	@echo "Testing settings GET..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test settings UPDATE (layout + convert_webp)
test-settings-update:
	@echo "Testing settings UPDATE..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Update settings (layout + convert_webp=false)..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"layout":{"primary_color":"#FF5733","secondary_color":"#33FF57","logo":"https://example.com/logo.png","theme":"Lara"},"convert_webp":false}'; \
	echo ""; echo "2. Read back settings..."; \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Authorization: Bearer $$TOKEN"
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
