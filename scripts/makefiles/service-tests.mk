.PHONY: test-service-create test-service-list test-service-update \
        test-service-delete test-services-all

# Test service create
test-service-create:
	@echo "Testing service create..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/services \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Corte Masculino","description":"Corte moderno","price":35.00,"duration":30,"is_active":true}'
	@echo ""

# Test service list
test-service-list:
	@echo "Testing service list..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/services \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test service update
test-service-update:
	@echo "Testing service update..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	SERVICE_ID=$$(curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/services \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/services/$$SERVICE_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Corte Premium","price":50.00}'
	@echo ""

# Test service delete
test-service-delete:
	@echo "Testing service delete..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	SERVICE_ID=$$(curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/services \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s -X DELETE http://localhost:8080/api/v1/$$URL_CODE/services/$$SERVICE_ID \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# E2E services
test-services-all:
	@echo "========================================="
	@echo "E2E: CRUD Services"
	@echo "========================================="
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Create:"; \
	RESPONSE=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/services \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Servico E2E","description":"Teste","price":25.00,"duration":60,"is_active":true}'); \
	echo "$$RESPONSE"; \
	SERVICE_ID=$$(echo "$$RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; echo "2. List:"; \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/services -H "Authorization: Bearer $$TOKEN"; \
	echo ""; echo "3. Update:"; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/services/$$SERVICE_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Servico E2E Atualizado","price":50.00}'; \
	echo ""; echo "4. Delete:"; \
	curl -s -X DELETE http://localhost:8080/api/v1/$$URL_CODE/services/$$SERVICE_ID \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""
	@echo "========================================="
