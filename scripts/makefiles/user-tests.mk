.PHONY: test-members-list test-members-can-add test-members-invite \
        test-roles-list test-app-users-list

# Test members list
test-members-list:
	@echo "Testing members list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/minha-loja/members \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test can-add member
test-members-can-add:
	@echo "Testing can-add member..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/minha-loja/members/can-add \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test invite member
test-members-invite:
	@echo "Testing invite member..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X POST http://localhost:8080/api/v1/minha-loja/members/invite \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"name":"Colaborador","email":"colab@minha-loja.com","password":"senha12345","role_slug":"member"}'
	@echo ""

# Test roles list
test-roles-list:
	@echo "Testing roles list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/minha-loja/roles \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""

# Test app users list
test-app-users-list:
	@echo "Testing app users list..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"joao@minha-loja.com","password":"senha12345"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
	curl -X GET http://localhost:8080/api/v1/minha-loja/app-users \
		-H "Authorization: Bearer $$TOKEN"
	@echo ""
