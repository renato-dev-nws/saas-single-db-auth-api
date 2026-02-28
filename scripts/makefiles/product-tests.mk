.PHONY: test-product-create test-product-list test-product-update \
        test-product-delete test-products-all

# Test product create
test-product-create:
	@echo "Testing product create..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Camiseta Basica","description":"Algodao 100%","price":49.90,"stock":100,"is_active":true}'
	@echo ""

# Test product list
test-product-list:
	@echo "Testing product list..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/products \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test product update
test-product-update:
	@echo "Testing product update..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	PRODUCT_ID=$$(curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/products \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Camiseta Premium","price":69.90}'
	@echo ""

# Test product delete
test-product-delete:
	@echo "Testing product delete..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	PRODUCT_ID=$$(curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/products \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s -X DELETE http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# E2E products
test-products-all:
	@echo "========================================="
	@echo "E2E: CRUD Products"
	@echo "========================================="
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "1. Create:"; \
	RESPONSE=$$(curl -s -X POST http://localhost:8080/api/v1/$$URL_CODE/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Produto E2E","description":"Teste","price":10.00,"stock":5,"is_active":true}'); \
	echo "$$RESPONSE"; \
	PRODUCT_ID=$$(echo "$$RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; echo "2. List:"; \
	curl -s -X GET http://localhost:8080/api/v1/$$URL_CODE/products -H "Authorization: Bearer $$TOKEN"; \
	echo ""; echo "3. Update:"; \
	curl -s -X PUT http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Produto E2E Atualizado","price":20.00}'; \
	echo ""; echo "4. Delete:"; \
	curl -s -X DELETE http://localhost:8080/api/v1/$$URL_CODE/products/$$PRODUCT_ID \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""
	@echo "========================================="
