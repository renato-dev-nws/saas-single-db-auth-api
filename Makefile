include scripts/makefiles/database.mk
include scripts/makefiles/admin-tests.mk
include scripts/makefiles/tenant-tests.mk
include scripts/makefiles/user-tests.mk
include scripts/makefiles/app-tests.mk
include scripts/makefiles/product-tests.mk
include scripts/makefiles/service-tests.mk
include scripts/makefiles/setting-tests.mk
include scripts/makefiles/security-tests.mk

.PHONY: up down build logs logs-admin logs-tenant logs-app dev-admin dev-tenant dev-app

# Docker compose
up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

logs-admin:
	docker compose logs -f admin-api

logs-tenant:
	docker compose logs -f tenant-api

logs-app:
	docker compose logs -f app-api

# Local dev (run outside Docker)
dev-admin:
	go run ./cmd/admin-api

dev-tenant:
	go run ./cmd/tenant-api

dev-app:
	go run ./cmd/app-api

# Build binaries
build-admin:
	go build -buildvcs=false -o bin/admin-api ./cmd/admin-api

build-tenant:
	go build -buildvcs=false -o bin/tenant-api ./cmd/tenant-api

build-app:
	go build -buildvcs=false -o bin/app-api ./cmd/app-api

build-all:
	@$(MAKE) build-admin
	@$(MAKE) build-tenant
	@$(MAKE) build-app

# Clean
clean:
	rm -rf bin/
	rm -rf uploads/*
	docker compose down -v

# Swagger
swagger:
	swag init -g cmd/admin-api/main.go -o docs/admin --parseDependency --parseInternal --exclude ./internal/handlers/tenant,./internal/handlers/app,./cmd/tenant-api,./cmd/app-api
	swag init -g cmd/tenant-api/main.go -o docs/tenant --parseDependency --parseInternal --exclude ./internal/handlers/admin,./internal/handlers/app,./cmd/admin-api,./cmd/app-api
	swag init -g cmd/app-api/main.go -o docs/app --parseDependency --parseInternal --exclude ./internal/handlers/admin,./internal/handlers/tenant,./cmd/admin-api,./cmd/tenant-api
