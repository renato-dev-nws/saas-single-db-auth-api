# SaaS Multi-Tenant API (Single Database)

API SaaS multi-tenant com isolamento lógico em banco de dados único.

## 🏗️ Arquitetura

- **3 APIs separadas**: Tenant (backoffice), Admin (gestão plataforma), App (usuários finais)
- **Multi-tenant**: Isolamento lógico via `tenant_id`
- **Stack**: Go 1.24, PostgreSQL 17, Redis, Docker

## 🚀 Quick Start

### 1. Iniciar ambiente

```bash
# Subir containers (PostgreSQL + Redis + 3 APIs)
make up

# Ver logs
make logs
```

### 2. Criar banco de dados e rodar migrations

```bash
# Rodar migrations (cria schema + seeds)
make db-migrate

# Verificar status
make db-status
```

**Seeds automáticos incluídos:**
- Admin: `admin@saas.com` / `admin123`
- 4 Planos: Starter, Business, Premium, Enterprise
- 2 Features: Products, Services
- 1 Promoção: 50% off por 3 meses

### 3. Testar APIs

```bash
# Tenant API (porta 8080)
make test-plans-public        # Listar planos
make test-subscription        # Criar tenant via public signup
make test-tenant-login        # Login backoffice
make test-testenovo           # E2E: criar tenant + config

# Admin API (porta 8081)
make test-admin-login         # Login admin
make test-plans-list          # Listar planos (admin)
make test-tenants-list        # Listar tenants
make test-promotions-list     # Listar promoções

# App API (porta 8082)
make test-app-register        # Registrar usuário app
make test-app-login           # Login app
make test-app-catalog         # Catálogo público
```

## 📦 Comandos Make

### Docker

```bash
make up              # Subir containers
make down            # Parar containers
make build           # Rebuild images
make logs            # Logs de todos os serviços
make logs-admin      # Logs Admin API
make logs-tenant     # Logs Tenant API
make logs-app        # Logs App API
```

### Database

```bash
make db-migrate      # Rodar migrations
make db-migrate-down # Rollback migrations
make db-reset        # Reset (down + up)
make db-recreate     # Drop + create + migrate
make db-status       # Status do banco
make db-backup       # Criar backup
make db-restore FILE=backups/xxx.sql  # Restaurar backup
make db-psql         # Abrir psql shell
make db-tables       # Listar tabelas
make db-tenants      # Listar tenants
make db-plans        # Listar planos
make db-admins       # Listar admins
```

### Testes

```bash
# Tenant API tests
make test-subscription        # Criar tenant (public)
make test-tenant-login        # Login backoffice
make test-user-me             # User info
make test-switch-tenant       # Trocar tenant ativo
make test-testenovo           # E2E completo

# Admin API tests
make test-admin-login         # Login admin
make test-sysusers-list       # Listar admin users
make test-plans-list          # Listar planos
make test-features-list       # Listar features
make test-tenants-list        # Listar tenants
make test-promotions-list     # Listar promoções

# User/Member tests
make test-members-list        # Listar membros
make test-members-invite      # Convidar membro
make test-roles-list          # Listar roles

# Product tests
make test-product-create      # Criar produto
make test-product-list        # Listar produtos
make test-products-all        # E2E CRUD products

# Service tests
make test-service-create      # Criar serviço
make test-service-list        # Listar serviços
make test-services-all        # E2E CRUD services

# App tests
make test-app-register        # Registrar app user
make test-app-login           # Login app
make test-app-catalog         # Catálogo público
make test-app-all             # E2E app completo

# Settings tests
make test-settings-list       # Listar settings
make test-settings-update     # Atualizar setting
```

### Dev Local

```bash
make dev-admin      # Rodar Admin API local
make dev-tenant     # Rodar Tenant API local
make dev-app        # Rodar App API local

make build-admin    # Build binário admin
make build-tenant   # Build binário tenant
make build-app      # Build binário app
make build-all      # Build todos
```

## 🔑 Credenciais Padrão

### Admin Login
- Email: `admin@saas.com`
- Password: `admin123`
- URL: `http://localhost:8081/api/v1/admin/auth/login`

### Tenant de Teste (após criar via subscription)
- Email: `joao@minha-loja.com`
- Password: `senha12345`
- URL Code: `minha-loja`
- URL: `http://localhost:8080/api/v1/auth/login`

## 🌐 Portas

- **8080**: Tenant API (backoffice)
- **8081**: Admin API (gestão plataforma)
- **8082**: App API (usuários finais)
- **5432**: PostgreSQL
- **6379**: Redis

## 📂 Estrutura

```
.
├── cmd/                    # Entry points (main.go)
│   ├── admin-api/
│   ├── tenant-api/
│   └── app-api/
├── internal/               # Código da aplicação
│   ├── domain/            # Entidades e interfaces
│   ├── repository/        # Acesso a dados
│   ├── service/           # Lógica de negócio
│   ├── handler/           # HTTP handlers
│   ├── middleware/        # Middlewares
│   └── util/              # Utilitários
├── migrations/            # SQL migrations
├── scripts/makefiles/     # Makefiles organizados
│   ├── database.mk
│   ├── admin-tests.mk
│   ├── tenant-tests.mk
│   └── ...
├── docker-compose.yml
├── Makefile              # Makefile principal
└── .env                  # Configurações
```

## 🔄 Workflow de Desenvolvimento

1. **Subir ambiente**: `make up`
2. **Migrations**: `make db-migrate`
3. **Verificar**: `make db-status`
4. **Testar**: `make test-admin-login`, `make test-subscription`, etc
5. **Ver logs**: `make logs-admin`, `make logs-tenant`
6. **Reset DB**: `make db-reset` (quando necessário)
7. **Parar**: `make down`

## 🐛 Debug

```bash
# Ver logs de uma API específica
make logs-admin
make logs-tenant
make logs-app

# Acessar shell do PostgreSQL
make db-psql

# Ver tabelas
make db-tables

# Ver tenants cadastrados
make db-tenants

# Backup antes de testar algo arriscado
make db-backup
```

## 📝 Notas

- Todos os testes rodam direto no WSL via curl + grep/cut (sem jq)
- Migrations já incluem seeds (admin, planos, features)
- Os 3 serviços (admin/tenant/app) rodam em containers separados
- Upload de arquivos vai para `./uploads` (bind mount)
- Backups são salvos em `./backups/`

## 🛠️ Comandos Úteis

```bash
# Rebuild completo
make down && make build && make up && make db-migrate

# Reset completo do banco
make db-recreate

# Criar backup antes de mudanças
make db-backup

# Restaurar backup
make db-restore FILE=backups/saasdb_20260225_120000.sql

# Limpar tudo (containers + volumes)
make clean
```
