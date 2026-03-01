.PHONY: test-security-cross-tenant test-validation-errors test-subscription-validation \
        test-subscription-company-validation test-security-all

# ─── Security: Cross-Tenant Access Test ──────────────────────
# Login as Joao (tenant HRP1ZYERFVA), then try to access another tenant's data
test-security-cross-tenant:
	@echo "========================================="
	@echo "SECURITY: Cross-Tenant Access Test"
	@echo "========================================="
	@echo ""
	@echo "1. Login as Joao (tenant Minha Loja)..."
	@LOGIN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}'); \
	TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	MY_URL=$$(echo "$$LOGIN" | grep -o '"current_tenant_code":"[^"]*' | cut -d'"' -f4); \
	echo "   Token obtained. My tenant: $$MY_URL"; \
	echo ""; \
	echo "2. Access OWN tenant bootstrap (should succeed)..."; \
	RESULT=$$(curl -s -o /dev/null -w '%{http_code}' -X GET http://localhost:8080/api/v1/$$MY_URL/bootstrap \
		-H "Authorization: Bearer $$TOKEN"); \
	if [ "$$RESULT" = "200" ]; then echo "   ✅ OWN tenant: HTTP $$RESULT (OK)"; else echo "   ❌ OWN tenant: HTTP $$RESULT (FAIL)"; fi; \
	echo ""; \
	echo "3. Get a DIFFERENT tenant url_code via admin API..."; \
	ADMIN_TOKEN=$$(curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@saas.com","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	OTHER_URL=$$(curl -s -X GET http://localhost:8081/api/v1/admin/tenants \
		-H "Authorization: Bearer $$ADMIN_TOKEN" | grep -o '"URLCode":"[^"]*' | cut -d'"' -f4 | grep -v "$$MY_URL" | head -1); \
	echo "   Other tenant: $$OTHER_URL"; \
	echo ""; \
	echo "4. Access OTHER tenant bootstrap with Joao's token (should be 403)..."; \
	RESULT=$$(curl -s -o /dev/null -w '%{http_code}' -X GET http://localhost:8080/api/v1/$$OTHER_URL/bootstrap \
		-H "Authorization: Bearer $$TOKEN"); \
	if [ "$$RESULT" = "403" ]; then echo "   ✅ OTHER tenant: HTTP $$RESULT (BLOCKED)"; else echo "   ❌ OTHER tenant: HTTP $$RESULT (VULNERABILITY!)"; fi; \
	echo ""; \
	echo "5. Error body for cross-tenant access:"; \
	curl -s -X GET http://localhost:8080/api/v1/$$OTHER_URL/bootstrap \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""
	@echo "========================================="

# ─── Validation: Error Format Test ───────────────────────────
# Send invalid payloads and check field-level error messages
test-validation-errors:
	@echo "========================================="
	@echo "VALIDATION: Error Format Test"
	@echo "========================================="
	@echo ""
	@echo "1. Empty subscription body (should return field errors)..."
	@curl -s -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{}'
	@echo ""
	@echo ""
	@echo "2. Invalid email in subscription..."
	@curl -s -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"name":"Test","email":"not-an-email","password":"123","plan_id":"invalid","billing_cycle":"monthly","subdomain":"test"}'
	@echo ""
	@echo ""
	@echo "3. Empty admin login body..."
	@curl -s -X POST http://localhost:8081/api/v1/admin/auth/login \
		-H "Content-Type: application/json" \
		-d '{}'
	@echo ""
	@echo ""
	@echo "4. Empty tenant login body..."
	@curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{}'
	@echo ""
	@echo "========================================="

# ─── Validation: Subscription company_name required when is_company=true ─────
test-subscription-company-validation:
	@echo "========================================="
	@echo "VALIDATION: company_name required when is_company=true"
	@echo "========================================="
	@echo ""
	@echo "1. is_company=true WITHOUT company_name (should fail)..."
	@curl -s -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"name":"Test User","email":"test-company@test.com","password":"senha12345","is_company":true,"plan_id":"10000000-0000-0000-0000-000000000001","billing_cycle":"monthly","subdomain":"testcomp"}'
	@echo ""
	@echo ""
	@echo "2. is_company=true WITH company_name (should succeed if email is new)..."
	@curl -s -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"name":"Test User","email":"test-company-ok@test.com","password":"senha12345","is_company":true,"company_name":"My Company","plan_id":"10000000-0000-0000-0000-000000000001","billing_cycle":"monthly","subdomain":"testcompok"}'
	@echo ""
	@echo ""
	@echo "3. is_company=false WITHOUT company_name (should succeed if email is new)..."
	@curl -s -X POST http://localhost:8080/api/v1/subscription \
		-H "Content-Type: application/json" \
		-d '{"name":"Individual User","email":"test-individual@test.com","password":"senha12345","is_company":false,"plan_id":"10000000-0000-0000-0000-000000000001","billing_cycle":"monthly","subdomain":"testindiv"}'
	@echo ""
	@echo "========================================="

# ─── Run all security and validation tests ───────────────────
test-security-all:
	@$(MAKE) test-security-cross-tenant
	@$(MAKE) test-validation-errors
	@$(MAKE) test-subscription-company-validation
