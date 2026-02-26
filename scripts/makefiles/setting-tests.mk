TENANT_URL=http://localhost:8080

# ─── Settings ───────────────────────────────────────────────
test-settings-list:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/settings" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-settings-get:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/settings/theme_color" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-settings-update:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/settings/theme_color" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"value":"#FF5733"}' | jq .
	@echo ""
