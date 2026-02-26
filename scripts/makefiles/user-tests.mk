TENANT_URL=http://localhost:8080

# ─── Membros ────────────────────────────────────────────────
test-members-list:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/members" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-members-can-add:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/members/can-add" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-members-invite:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/members/invite" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Colaborador","email":"colab@minha-loja.com","password":"senha12345","role_slug":"member"}' | jq .
	@echo ""

# ─── Roles ──────────────────────────────────────────────────
test-roles-list:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/roles" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

# ─── App Users ──────────────────────────────────────────────
test-app-users-list:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/app-users" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""
