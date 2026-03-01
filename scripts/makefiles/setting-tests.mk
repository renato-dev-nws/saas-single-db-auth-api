.PHONY: test-settings-get test-settings-update \
        test-layout-get test-layout-update \
        test-language test-language-all

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

# Test settings UPDATE (layout + convert_webp + language)
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
	echo ""; echo "2. Update language to en via settings PUT..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"language":"en"}'; \
	echo ""; echo "3. Read back settings (language=en)..."; \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Authorization: Bearer $$TOKEN"; \
	echo ""; echo "4. Revert language to pt-BR..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"language":"pt-BR"}'; \
	echo ""; echo "5. Read back settings (language=pt-BR)..."; \
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

# Test language switch (en → es → pt-BR round-trip)
test-language:
	@echo "Testing language switch..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Change language to en..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings/language \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"language":"en"}'; \
	echo ""; echo "2. Get settings (language=en)..."; \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Authorization: Bearer $$TOKEN"; \
	echo ""; echo "3. Invalid language (should fail in English)..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings/language \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"language":"xyz"}'; \
	echo ""; echo "4. Change language to es..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings/language \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"language":"es"}'; \
	echo ""; echo "5. Invalid language (should fail in Spanish)..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings/language \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"language":"xyz"}'; \
	echo ""; echo "6. Change back to pt-BR..."; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/settings/language \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"language":"pt-BR"}'; \
	echo ""; echo "7. Verify settings (language=pt-BR)..."; \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/settings \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Run all language tests
test-language-all: test-language
