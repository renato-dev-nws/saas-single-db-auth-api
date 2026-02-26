APP_URL=http://localhost:8082
URL_CODE=minha-loja

# ─── Registro + Login ───────────────────────────────────────
test-app-register:
	@echo "Registrando app user..."
	@curl -s -X POST $(APP_URL)/api/v1/$(URL_CODE)/auth/register \
		-H "Content-Type: application/json" \
		-d '{"name":"Cliente App","email":"cliente@app.com","password":"senha12345"}' | jq .
	@echo ""

test-app-login:
	@curl -s -X POST $(APP_URL)/api/v1/$(URL_CODE)/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"cliente@app.com","password":"senha12345"}' | jq .
	@echo ""

# ─── Catálogo (público) ─────────────────────────────────────
test-app-catalog:
	@echo "Produtos públicos:"
	@curl -s "$(APP_URL)/api/v1/$(URL_CODE)/catalog/products" | jq .
	@echo ""
	@echo "Serviços públicos:"
	@curl -s "$(APP_URL)/api/v1/$(URL_CODE)/catalog/services" | jq .
	@echo ""

# ─── Perfil (autenticado) ───────────────────────────────────
test-app-profile:
	@LOGIN=$$(curl -s -X POST $(APP_URL)/api/v1/$(URL_CODE)/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"cliente@app.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	echo "Perfil:"; \
	curl -s "$(APP_URL)/api/v1/$(URL_CODE)/profile" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

# ─── Teste E2E App ──────────────────────────────────────────
test-app-all:
	@echo "========================================="
	@echo "Teste E2E: App User Completo"
	@echo "========================================="
	@echo ""
	@echo "1. Registrar..."
	@REGISTER=$$(curl -s -X POST $(APP_URL)/api/v1/$(URL_CODE)/auth/register \
		-H "Content-Type: application/json" \
		-d '{"name":"E2E User","email":"e2e@app.com","password":"senha12345"}'); \
	echo "$$REGISTER" | jq .; \
	TOKEN=$$(echo "$$REGISTER" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	echo ""; \
	echo "2. Me..."; \
	curl -s "$(APP_URL)/api/v1/$(URL_CODE)/auth/me" \
		-H "Authorization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "3. Catálogo..."; \
	curl -s "$(APP_URL)/api/v1/$(URL_CODE)/catalog/products" | jq .; \
	echo ""; \
	echo "4. Perfil..."; \
	curl -s "$(APP_URL)/api/v1/$(URL_CODE)/profile" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""
	@echo "========================================="
	@echo "Teste concluído!"
	@echo "========================================="
