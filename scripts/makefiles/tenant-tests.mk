TENANT_URL=http://localhost:8080

# ─── Subscription (Público) ────────────────────────────────
test-subscription:
	@echo "Criando novo tenant via subscription pública..."
	@curl -s -X POST $(TENANT_URL)/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","tenant_name":"Minha Loja","url_code":"minha-loja","is_company":false,"owner_name":"João Silva","owner_email":"joao@minha-loja.com","owner_password":"senha12345"}' | jq .
	@echo ""

test-subscription-with-promo:
	@echo "Criando tenant com promoção..."
	@curl -s -X POST $(TENANT_URL)/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","promotion_id":"pppppppp-pppp-pppp-pppp-pppppppppppp","tenant_name":"Loja Promo","url_code":"loja-promo","is_company":false,"owner_name":"Maria Promo","owner_email":"maria@loja-promo.com","owner_password":"senha12345"}' | jq .
	@echo ""

test-plans-public:
	@curl -s $(TENANT_URL)/api/v1/plans | jq .
	@echo ""

# ─── Auth Backoffice ───────────────────────────────────────
test-user-login:
	@curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | jq .
	@echo ""

test-user-me:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -s $(TENANT_URL)/api/v1/auth/me \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

test-tenant-config:
	@LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/config" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""

# ─── Teste E2E Completo ────────────────────────────────────
test-new-tenant:
	@echo "========================================="
	@echo "Teste E2E: Criar Tenant + Login + Config"
	@echo "========================================="
	@echo ""
	@echo "1. Criando tenant..."
	@RESPONSE=$$(curl -s -X POST $(TENANT_URL)/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","tenant_name":"Nova Empresa","url_code":"nova-empresa","is_company":false,"owner_name":"Novo Usuario","owner_email":"novo@empresa.com","owner_password":"senha12345"}'); \
	echo "$$RESPONSE" | jq .; \
	TOKEN=$$(echo "$$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	URL_CODE=$$(echo "$$RESPONSE" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
	echo ""; \
	echo "2. Config do tenant $$URL_CODE..."; \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/config" \
		-H "Authorization: Bearer $$TOKEN" | jq .; \
	echo ""; \
	echo "3. Listando produtos (feature check)..."; \
	curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
		-H "Authorization: Bearer $$TOKEN" | jq .
	@echo ""
	@echo "========================================="
	@echo "Teste concluído!"
	@echo "========================================="
