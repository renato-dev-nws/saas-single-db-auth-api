.PHONY: test-app-register test-app-login test-app-catalog test-app-profile test-app-all

# Test app user register
test-app-register:
	@echo "Testing app user register..."
	@curl -X POST http://localhost:8082/api/v1/minha-loja/auth/register \
		-H "Content-Type: application/json" \
		-d '{"name":"Cliente App","email":"cliente@app.com","password":"senha12345"}'
	@echo ""

# Test app user login
test-app-login:
	@echo "Testing app user login..."
	@curl -X POST http://localhost:8082/api/v1/minha-loja/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"cliente@app.com","password":"senha12345"}'
	@echo ""

# Test catalog (public)
test-app-catalog:
	@echo "Testing catalog products..."
	@curl -X GET http://localhost:8082/api/v1/minha-loja/catalog/products
	@echo ""
	@echo "Testing catalog services..."
	@curl -X GET http://localhost:8082/api/v1/minha-loja/catalog/services
	@echo ""

# Test profile (protected)
test-app-profile:
	@echo "Testing app profile..."
	@TOKEN=$$(curl -s -X POST http://localhost:8082/api/v1/minha-loja/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"cliente@app.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8082/api/v1/minha-loja/profile \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# E2E app
test-app-all:
	@echo "========================================="
	@echo "E2E: App User Complete"
	@echo "========================================="
	@echo "1. Register:"
	@RESPONSE=$$(curl -s -X POST http://localhost:8082/api/v1/minha-loja/auth/register \
		-H "Content-Type: application/json" \
		-d '{"name":"E2E User","email":"e2e@app.com","password":"senha12345"}'); \
	echo "$$RESPONSE"; \
	TOKEN=$$(echo "$$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	echo ""; echo "2. Me:"; \
	curl -s -X GET http://localhost:8082/api/v1/minha-loja/auth/me \
		-H "Authorization: Bearer $$TOKEN"; \
	echo ""; echo "3. Catalog:"; \
	curl -s -X GET http://localhost:8082/api/v1/minha-loja/catalog/products; \
	echo ""; echo "4. Profile:"; \
	curl -s -X GET http://localhost:8082/api/v1/minha-loja/profile \
		-H "Authorization: Bearer $$TOKEN"
	@echo "========================================="
