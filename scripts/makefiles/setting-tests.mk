.PHONY: test-settings-list test-settings-get test-settings-update

# Test settings list
test-settings-list:
	@echo "Testing settings list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/minha-loja/settings \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test setting get
test-settings-get:
	@echo "Testing setting get..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/minha-loja/settings/theme_color \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test setting update
test-settings-update:
	@echo "Testing setting update..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X PUT http://localhost:8080/api/v1/minha-loja/settings/theme_color \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"value":"#FF5733"}'
	@echo ""
