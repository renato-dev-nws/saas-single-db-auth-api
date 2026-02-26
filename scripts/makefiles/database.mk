.PHONY: db-migrate db-migrate-down db-reset db-recreate db-status db-backup db-restore \
        db-psql db-logs db-shell

# Database migration commands
db-migrate:
	@echo "Running migrations..."
	@docker compose exec postgres psql -U saasuser -d saasdb -f /migrations/001_initial_schema.up.sql
	@echo "✓ Migrations completed"

db-migrate-down:
	@echo "Rolling back migrations..."
	@docker compose exec postgres psql -U saasuser -d saasdb -f /migrations/001_initial_schema.down.sql
	@echo "✓ Rollback completed"

# Reset database (down + up)
db-reset:
	@echo "Resetting database..."
	@$(MAKE) db-migrate-down
	@$(MAKE) db-migrate
	@echo "✓ Database reset complete"

# Recreate database from scratch
db-recreate:
	@echo "Recreating database..."
	@docker compose exec postgres psql -U saasuser -d postgres -c "DROP DATABASE IF EXISTS saasdb;"
	@docker compose exec postgres psql -U saasuser -d postgres -c "CREATE DATABASE saasdb OWNER saasuser;"
	@$(MAKE) db-migrate
	@echo "✓ Database recreated"

# Check database status
db-status:
	@echo "Database status:"
	@docker compose exec postgres psql -U saasuser -d saasdb -c "\
		SELECT 'Tables: ' || COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'; \
		SELECT 'Tenants: ' || COUNT(*) FROM tenants; \
		SELECT 'Users: ' || COUNT(*) FROM users; \
		SELECT 'Admin Users: ' || COUNT(*) FROM system_admin_users;"

# Backup database
db-backup:
	@echo "Creating backup..."
	@mkdir -p backups
	@docker compose exec -T postgres pg_dump -U saasuser saasdb > backups/saasdb_$$(date +%Y%m%d_%H%M%S).sql
	@echo "✓ Backup created in backups/"

# Restore from backup (usage: make db-restore FILE=backups/saasdb_20260225_120000.sql)
db-restore:
	@if [ -z "$(FILE)" ]; then \
		echo "Error: FILE parameter required. Usage: make db-restore FILE=backups/saasdb_20260225_120000.sql"; \
		exit 1; \
	fi
	@echo "Restoring from $(FILE)..."
	@docker compose exec -T postgres psql -U saasuser -d saasdb < $(FILE)
	@echo "✓ Database restored"

# Open psql shell
db-psql:
	@docker compose exec postgres psql -U saasuser -d saasdb

# Show database logs
db-logs:
	@docker compose logs -f postgres

# Open bash shell in postgres container
db-shell:
	@docker compose exec postgres bash

# Show all tables
db-tables:
	@docker compose exec postgres psql -U saasuser -d saasdb -c "\dt"

# Show all tenants
db-tenants:
	@docker compose exec postgres psql -U saasuser -d saasdb -c "SELECT id, name, url_code, status, created_at FROM tenants;"

# Show all plans
db-plans:
	@docker compose exec postgres psql -U saasuser -d saasdb -c "SELECT id, name, price, max_users, is_active FROM plans;"

# Show all admin users
db-admins:
	@docker compose exec postgres psql -U saasuser -d saasdb -c "SELECT id, name, email, status FROM system_admin_users;"
