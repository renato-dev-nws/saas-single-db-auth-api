.PHONY: test-product-create test-product-list test-product-update \
        test-product-delete test-products-all

# Test product create
test-product-create:
	@echo "Testing product create..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8080/api/v1/minha-loja/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Camiseta Basica","description":"Algodao 100%","price":49.90,"stock":100,"is_active":true}'
	@echo ""

# Test product list
test-product-list:
	@echo "Testing product list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/minha-loja/products \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test product update
test-product-update:
	@echo "Testing product update..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	PRODUCT_ID=$$(curl -s -X GET http://localhost:8080/api/v1/minha-loja/products \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -X PUT http://localhost:8080/api/v1/minha-loja/products/$$PRODUCT_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Camiseta Premium","price":69.90}'
	@echo ""

# Test product delete
test-product-delete:
	@echo "Testing product delete..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	PRODUCT_ID=$$(curl -s -X GET http://localhost:8080/api/v1/minha-loja/products \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -X DELETE http://localhost:8080/api/v1/minha-loja/products/$$PRODUCT_ID \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# E2E products
test-products-all:
	@echo "========================================="
	@echo "E2E: CRUD Products"
	@echo "========================================="
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	echo "1. Create:"; \
	RESPONSE=$$(curl -s -X POST http://localhost:8080/api/v1/minha-loja/products \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Produto E2E","description":"Teste","price":10.00,"stock":5,"is_active":true}'); \
	echo "$$RESPONSE"; \
	PRODUCT_ID=$$(echo "$$RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; echo "2. List:"; \
	curl -s -X GET http://localhost:8080/api/v1/minha-loja/products -H "Authorization: Bearer $$TOKEN"; \
	echo ""; echo "3. Update:"; \
	curl -s -X PUT http://localhost:8080/api/v1/minha-loja/products/$$PRODUCT_ID \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Produto E2E Atualizado","price":20.00}'; \
	echo ""; echo "4. Delete:"; \
	curl -s -X DELETE http://localhost:8080/api/v1/minha-loja/products/$$PRODUCT_ID \
		-H "Authorization: Bearer $$TOKEN"
	@echo "========================================="
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	cul -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authoization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Camiseta Básica","slug":"camiseta-basica","desciption":"Algodão 100%","price":49.90,"stock_quantity":100,"is_active":true}' | jq .
	@echo ""

test-poduct-list:
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authoization: Bearer $$TOKEN" | jq .
	@echo ""

test-poduct-update:
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	PRODUCT_ID=$$(cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authoization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "Atualizando poduto $$PRODUCT_ID"; \
	cul -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authoization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Camiseta Pemium","price":69.90}' | jq .
	@echo ""

test-poduct-delete:
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	PRODUCT_ID=$$(cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authoization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "Deletando poduto $$PRODUCT_ID"; \
	cul -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authoization: Bearer $$TOKEN" | jq .
	@echo ""

# ─── Teste E2E Podutos ─────────────────────────────────────
test-poducts-all:
	@echo "========================================="
	@echo "Teste E2E: CRUD Podutos"
	@echo "========================================="
	@echo ""
	@LOGIN=$$(cul -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","passwod":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | gep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | gep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "1. Ciar produto..."; \
	CREATE=$$(cul -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authoization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Poduto E2E","slug":"produto-e2e","description":"Teste E2E","price":10.00,"stock_quantity":5,"is_active":true}'); \
	echo "$$CREATE" | jq .; \
	PRODUCT_ID=$$(echo "$$CREATE" | gep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; \
	echo "2. Lista produtos..."; \
	cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authoization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "3. Obte produto $$PRODUCT_ID..."; \
	cul -s "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authoization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "4. Atualiza produto..."; \
	cul -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authoization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Poduto E2E Atualizado","price":20.00}' | jq .; \
	echo ""; \
	echo "5. Deleta produto..."; \
	cul -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authoization: Bearer $$TOKEN" | jq .
	@echo ""
	@echo "========================================="
	@echo "Teste concluído!"
	@echo "========================================="
