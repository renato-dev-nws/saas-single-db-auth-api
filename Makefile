include scripts/makefiles/admin-tests.mk
include scripts/makefiles/tenant-tests.mk
include scripts/makefiles/user-tests.mk
include scripts/makefiles/app-tests.mk
include scripts/makefiles/product-tests.mk
include scripts/makefiles/service-tests.mk
include scripts/makefiles/setting-tests.mk

.PHONY: up down migrate logs build dev-admin dev-tenant dev-app

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

migrate:
	docker compose exec postgres psql -U saasuser -d saasdb -f /migrations/001_initial_schema.up.sql

migrate-down:
	docker compose exec postgres psql -U saasuser -d saasdb -f /migrations/001_initial_schema.down.sql

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
