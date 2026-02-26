.PHONY: test-service-create test-service-list test-service-update \
        test-service-delete test-services-all

# Test service create
test-service-create:
	@echo "Testing service create..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8080/api/v1/minha-loja/services \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Corte Masculino","description":"Corte moderno","price":35.00,"duration":30,"is_active":true}'
	@echo ""

# Test service list
test-service-list:
	@echo "Testing service list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/minha-loja/services \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test service update
test-service-update:
	@echo "Testing service update..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	SERVICE_ID=$$(curl -s -X GET http://localhost:8080/api/v1/minha-loja/services \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -X PUT http://localhost:8080/api/v1/minha-loja/services/$$SERVICE_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Corte Premium","price":50.00}'
	@echo ""

# Test service delete
test-service-delete:
	@echo "Testing service delete..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	SERVICE_ID=$$(curl -s -X GET http://localhost:8080/api/v1/minha-loja/services \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -X DELETE http://localhost:8080/api/v1/minha-loja/services/$$SERVICE_ID \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# E2E services
test-services-all:
	@echo "========================================="
	@echo "E2E: CRUD Services"
	@echo "========================================="
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	echo "1. Create:"; \
	RESPONSE=$$(curl -s -X POST http://localhost:8080/api/v1/minha-loja/services \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Servico E2E","description":"Teste","price":25.00,"duration":60,"is_active":true}'); \
	echo "$$RESPONSE"; \
	SERVICE_ID=$$(echo "$$RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; echo "2. List:"; \
	curl -s -X GET http://localhost:8080/api/v1/minha-loja/services -H "Authorization: Bearer $$TOKEN"; \
	echo ""; echo "3. Update:"; \
	curl -s -X PUT http://localhost:8080/api/v1/minha-loja/services/$$SERVICE_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Servico E2E Atualizado","price":50.00}'; \
	echo ""; echo "4. Delete:"; \
	curl -s -X DELETE http://localhost:8080/api/v1/minha-loja/services/$$SERVICE_ID \
		-H "Authorization: Bearer $$TOKEN"
	@echo "========================================="
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	cul -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authoization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Cote Masculino","slug":"corte-masculino","description":"Corte moderno","price":35.00,"duration_minutes":30,"is_active":true}' | jq .
	@echo ""

test-sevice-list:
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authoization: Bearer $$TOKEN" | jq .
	@echo ""

test-sevice-update:
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	SERVICE_ID=$$(cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authoization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "Atualizando seviço $$SERVICE_ID"; \
	cul -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authoization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Cote Premium","price":50.00,"duration_minutes":45}' | jq .
	@echo ""

test-sevice-delete:
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	SERVICE_ID=$$(cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authoization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "Deletando seviço $$SERVICE_ID"; \
	cul -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authoization: Bearer $$TOKEN" | jq .
	@echo ""

# ─── Teste E2E Seviços ────────────────────────────────────
test-sevices-all:
	@echo "========================================="
	@echo "Teste E2E: CRUD Seviços"
	@echo "========================================="
	@echo ""
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "1. Ciar serviço..."; \
	CREATE=$$(cul -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authoization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Seviço E2E","slug":"servico-e2e","description":"Teste E2E","price":25.00,"duration_minutes":60,"is_active":true}'); \
	echo "$$CREATE" | jq .; \
	SERVICE_ID=$$(echo "$$CREATE" | gep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; \
	echo "2. Lista serviços..."; \
	cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/services" \
		-H "Authoization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "3. Obte serviço $$SERVICE_ID..."; \
	cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authoization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "4. Atualiza serviço..."; \
	cul -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authoization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Seviço E2E Atualizado","price":50.00}' | jq .; \
	echo ""; \
	echo "5. Deleta serviço..."; \
	cul -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/services/$$SERVICE_ID" \
		-H "Authoization: Bearer $$TOKEN" | jq .
	@echo ""
	@echo "========================================="
	@echo "Teste concluído!"
	@echo "========================================="
