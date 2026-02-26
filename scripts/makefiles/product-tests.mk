TENANT_URL=http://localhost:8080

# ─── Produtos ───────────────────────────────────────────────
test-product-create:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Camiseta Básica","slug":"camiseta-basica","description":"Algodão 100%","price":49.90,"stock_quantity":100,"is_active":true}' | jq .
	@echo ""

test-product-list:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-product-update:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	PRODUCT_ID=$$(curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "Atualizando produto $$PRODUCT_ID"; \
	curl -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Camiseta Premium","price":69.90}' | jq .
	@echo ""

test-product-delete:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	PRODUCT_ID=$$(curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authorization: Bearer $$TOKEN" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "Deletando produto $$PRODUCT_ID"; \
	curl -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

# ─── Teste E2E Produtos ─────────────────────────────────────
test-products-all:
	@echo "========================================="
	@echo "Teste E2E: CRUD Produtos"
	@echo "========================================="
	@echo ""
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo "1. Criar produto..."; \
	CREATE=$$(curl -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Produto E2E","slug":"produto-e2e","description":"Teste E2E","price":10.00,"stock_quantity":5,"is_active":true}'); \
	echo "$$CREATE" | jq .; \
	PRODUCT_ID=$$(echo "$$CREATE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; \
	echo "2. Listar produtos..."; \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authorization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "3. Obter produto $$PRODUCT_ID..."; \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authorization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "4. Atualizar produto..."; \
	curl -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"Produto E2E Atualizado","price":20.00}' | jq .; \
	echo ""; \
	echo "5. Deletar produto..."; \
	curl -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PRODUCT_ID" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""
	@echo "========================================="
	@echo "Teste concluído!"
	@echo "========================================="
