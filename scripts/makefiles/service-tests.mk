TENANT_URL=http://localhost:8080

# ─── Serviços ───────────────────────────────────────────────
test-service-create:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Corte Masculino","slug":"corte-masculino","description":"Corte moderno","price":35.00,"duration_minutes":30,"is_active":true}' | jq .
	@echo ""

test-service-list:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-service-update:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	SERVICE_ID=$$(curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "Atualizando serviço $$SERVICE_ID"; \
	curl -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Corte Premium","price":50.00,"duration_minutes":45}' | jq .
	@echo ""

test-service-delete:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	SERVICE_ID=$$(curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "Deletando serviço $$SERVICE_ID"; \
	curl -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

# ─── Teste E2E Serviços ────────────────────────────────────
test-services-all:
	@echo "========================================="
	@echo "Teste E2E: CRUD Serviços"
	@echo "========================================="
	@echo ""
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "1. Criar serviço..."; \
	CREATE=$$(curl -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Serviço E2E","slug":"servico-e2e","description":"Teste E2E","price":25.00,"duration_minutes":60,"is_active":true}'); \
	echo "$$CREATE" | jq .; \
	SERVICE_ID=$$(echo "$$CREATE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; \
	echo "2. Listar serviços..."; \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authorization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "3. Obter serviço $$SERVICE_ID..."; \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authorization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "4. Atualizar serviço..."; \
	curl -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Serviço E2E Atualizado","price":50.00}' | jq .; \
	echo ""; \
	echo "5. Deletar serviço..."; \
	curl -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""
	@echo "========================================="
	@echo "Teste concluído!"
	@echo "========================================="
