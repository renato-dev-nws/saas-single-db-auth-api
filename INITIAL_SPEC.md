# Especificação Completa — SaaS Multi-Tenant (Single Database, Logical Isolation)

> Documento de instrução para geração de projeto em branco via GitHub Copilot.  
> Baseado na arquitetura do projeto `saas-multi-database-api`, adaptado para **banco único** com isolamento lógico por `tenant_id`.

---

## 1. Visão Geral

API SaaS em **Go** com **isolamento lógico de tenants** via coluna `tenant_id` em um único banco PostgreSQL. O sistema é dividido em três serviços independentes (processos/portas distintas):

| Serviço       | Porta | Audiência                                    |
|---------------|-------|----------------------------------------------|
| `tenant-api`  | 8080  | Usuários e donos de tenants (backoffice)     |
| `admin-api`   | 8081  | Administradores do SaaS (sistema)            |
| `app-api`     | 8082  | Usuários/clientes finais dos tenants         |

---

## 2. Stack Tecnológica

- **Linguagem:** Go 1.22+
- **Framework HTTP:** Gin (`github.com/gin-gonic/gin`)
- **Banco de Dados:** PostgreSQL 16 (Driver: `pgx/v5` com `pgxpool.Pool`)
- **Cache:** Redis 7 (`github.com/redis/go-redis/v9`)
- **Autenticação:** JWT (`github.com/golang-jwt/jwt/v5`)
- **Hashing:** `bcrypt` com custo 12
- **Armazenamento:** Local / AWS S3 / Cloudflare R2 (interface única via `storage.Provider`)
- **Migrações:** `golang-migrate/migrate` ou `pressly/goose`
- **Infraestrutura:** Docker + Docker Compose
- **Testes de API:** `curl` via targets do `Makefile`

---

## 3. Tipos de Usuário

### 3.1 `system_admin_users` — Administradores do SaaS
Equipe interna da plataforma. Gerencia tenants, planos, features e outros admins.  
**Tabelas:** `system_admin_users`, `system_admin_profiles`, `system_admin_roles`, `system_admin_permissions`, `system_admin_user_roles`, `system_admin_role_permissions`

### 3.2 `users` — Usuários do Tenant (Backoffice)
Donos e colaboradores de um tenant. Acessam o backoffice para gerenciar produtos, serviços, configurações e usuários do app.  
**Tabelas:** `users`, `user_profiles`, `user_roles`, `user_permissions`, `user_role_permissions`, `tenant_members`

### 3.3 `tenant_app_users` — Usuários/Clientes do App do Tenant
Usuários finais cadastrados no site/app do tenant (clientes, consumidores).  
**Tabelas:** `tenant_app_users`, `tenant_app_user_profiles`

---

## 4. Schema do Banco de Dados (PostgreSQL)

### 4.1 Extensões
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- Para buscas full-text eficientes
```

### 4.2 Enums
```sql
CREATE TYPE tenant_status    AS ENUM ('active', 'suspended', 'cancelled');
CREATE TYPE billing_cycle    AS ENUM ('monthly', 'quarterly', 'semiannual', 'annual');
CREATE TYPE user_status      AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE storage_provider AS ENUM ('local', 's3', 'r2');
CREATE TYPE discount_type    AS ENUM ('percent', 'fixed');
```

> **Soft Delete:** Todas as tabelas de entidades mutáveis possuem coluna `deleted_at TIMESTAMP NULL`.  
> Registros com `deleted_at IS NOT NULL` são tratados como deletados. Queries de listagem e validações **sempre** filtram `WHERE deleted_at IS NULL`.  
> O campo `status` (`inactive`, `suspended`) é usado para suspensão operacional sem exclusão de dados. O `deleted_at` é para remoção definitiva mas reversível.

---

### 4.3 Tabelas dos Administradores do SaaS

```sql
-- Administradores da plataforma SaaS
CREATE TABLE system_admin_users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255),
    email         VARCHAR(255) UNIQUE NOT NULL,
    hash_pass     VARCHAR(255) NOT NULL,
    status        user_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP    -- soft delete
);

CREATE TABLE system_admin_profiles (
    admin_user_id UUID         PRIMARY KEY REFERENCES system_admin_users(id) ON DELETE CASCADE,
    full_name     VARCHAR(255),
    title         VARCHAR(255),
    bio           TEXT,
    avatar_url    TEXT,
    social_links  JSONB        NOT NULL DEFAULT '{}',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE system_admin_roles (
    id          SERIAL       PRIMARY KEY,
    title       VARCHAR(100) NOT NULL,
    slug        VARCHAR(50)  UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE system_admin_permissions (
    id          SERIAL       PRIMARY KEY,
    title       VARCHAR(100) NOT NULL,
    slug        VARCHAR(50)  UNIQUE NOT NULL,
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE system_admin_user_roles (
    admin_user_id UUID    REFERENCES system_admin_users(id) ON DELETE CASCADE,
    admin_role_id INTEGER REFERENCES system_admin_roles(id) ON DELETE CASCADE,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (admin_user_id, admin_role_id)
);

CREATE TABLE system_admin_role_permissions (
    admin_role_id       INTEGER REFERENCES system_admin_roles(id) ON DELETE CASCADE,
    admin_permission_id INTEGER REFERENCES system_admin_permissions(id) ON DELETE CASCADE,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (admin_role_id, admin_permission_id)
);
```

---

### 4.4 Tabelas de Planos e Features

```sql
CREATE TABLE features (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    title       VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    code        VARCHAR(10)  UNIQUE NOT NULL, -- ex: 'prod', 'serv', 'blog'
    description TEXT,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE plans (
    id            UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255)   NOT NULL,
    description   TEXT,
    price         DECIMAL(10,2)  NOT NULL DEFAULT 0,
    max_users     INTEGER        NOT NULL DEFAULT 1, -- 1=owner only, 3=owner+2 colaboradores, 5=owner+4, etc.
    is_multilang  BOOLEAN        NOT NULL DEFAULT false, -- Habilita suporte a múltiplos idiomas no site do tenant
    is_active     BOOLEAN        NOT NULL DEFAULT true,
    created_at    TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE TABLE plan_features (
    plan_id    UUID REFERENCES plans(id)    ON DELETE CASCADE,
    feature_id UUID REFERENCES features(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (plan_id, feature_id)
);
```

---

### 4.5 Tabelas de Tenants

```sql
-- Promoções disponíveis para aplicar em contratações
CREATE TABLE promotions (
    id              UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255)   NOT NULL,
    description     TEXT,
    discount_type   discount_type  NOT NULL,         -- 'percent' ou 'fixed'
    discount_value  DECIMAL(10,2)  NOT NULL,         -- % (ex: 50.00) ou valor fixo (ex: 100.00)
    duration_months INTEGER        NOT NULL DEFAULT 1, -- duração do desconto em meses
    valid_from      TIMESTAMP      NOT NULL DEFAULT NOW(), -- promoção válida para novos cadastros a partir de
    valid_until     TIMESTAMP,                       -- NULL = sem prazo para novos cadastros
    is_active       BOOLEAN        NOT NULL DEFAULT true,
    created_at      TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE TABLE tenants (
    id            UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255)   NOT NULL,            -- nome de exibição do tenant
    url_code      VARCHAR(20)    UNIQUE NOT NULL,      -- ex: 'minha-loja', 'empresa-x'
    subdomain     VARCHAR(50)    UNIQUE NOT NULL,
    is_company    BOOLEAN        NOT NULL DEFAULT false,
    company_name  VARCHAR(255),                       -- preenchido se is_company = true
    custom_domain VARCHAR(255),                       -- domínio customizado (ex: app.empresa.com)
    status        tenant_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP      NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP                           -- soft delete
);

-- Histórico de contratações de plano do tenant (1 registro ativo por tenant)
CREATE TABLE tenant_plans (
    id                UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id         UUID           NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_id           UUID           NOT NULL REFERENCES plans(id),
    billing_cycle     billing_cycle  NOT NULL DEFAULT 'monthly',
    base_price        DECIMAL(10,2)  NOT NULL,  -- preço da tabela do plano no momento da contratação
    contracted_price  DECIMAL(10,2)  NOT NULL,  -- preço efetivamente cobrado (com ou sem promo)
    price_updated_at  TIMESTAMP      NOT NULL DEFAULT NOW(), -- data do último ajuste de valor
    promotion_id      UUID           REFERENCES promotions(id) ON DELETE SET NULL,
    promo_price       DECIMAL(10,2), -- preço durante o período promocional (calculado na criação)
    promo_expires_at  TIMESTAMP,     -- quando a promoção vence para este tenant (NULL = sem promo)
    is_active         BOOLEAN        NOT NULL DEFAULT true,  -- apenas 1 registro ativo por tenant
    started_at        TIMESTAMP      NOT NULL DEFAULT NOW(),
    ended_at          TIMESTAMP,     -- preenchido quando substituído por novo plano
    created_at        TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP      NOT NULL DEFAULT NOW()
);

-- Índice para garantir apenas 1 contrato ativo por tenant
CREATE UNIQUE INDEX idx_tenant_plans_active ON tenant_plans(tenant_id) WHERE is_active = true;

CREATE TABLE tenant_profiles (
    tenant_id       UUID         PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    about           TEXT,
    logo_url        TEXT,
    custom_settings JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW()
);
```

> **Lógica de preço em vigor:**  
> - Se `promo_expires_at IS NOT NULL AND promo_expires_at > NOW()` → cobrar `promo_price`  
> - Caso contrário → cobrar `contracted_price`  
> - O campo `contracted_price` é atualizado (`price_updated_at`) quando o admin reajusta o valor fora do contexto promocional.

---

### 4.6 Tabelas de Usuários do Tenant (Backoffice)

```sql
-- Usuários do backoffice (donos, colaboradores)
CREATE TABLE users (
    id                   UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                 VARCHAR(255)   NOT NULL, 
    email                VARCHAR(255) UNIQUE NOT NULL,
    hash_pass            VARCHAR(255) NOT NULL,
    last_tenant_url_code VARCHAR(20),  -- Último tenant acessado (para redirect no login)
    status               user_status  NOT NULL DEFAULT 'active',
    created_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMP    -- soft delete
);

CREATE TABLE user_profiles (
    user_id    UUID         PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    full_name  VARCHAR(255),
    about      TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- Roles são por tenant (tenant_id NULL = role global/template, copiada ao criar tenant)
CREATE TABLE user_roles (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id  UUID         REFERENCES tenants(id) ON DELETE CASCADE,  -- NULL = global template
    title      VARCHAR(255) NOT NULL,
    slug       VARCHAR(100) NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, slug)
);

CREATE TABLE user_permissions (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    feature_id  UUID         REFERENCES features(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL, -- ex: 'prod_c', 'prod_r', 'serv_d'
    description TEXT,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE user_role_permissions (
    role_id       UUID REFERENCES user_roles(id)       ON DELETE CASCADE,
    permission_id UUID REFERENCES user_permissions(id)  ON DELETE CASCADE,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- Associação usuário <-> tenant (um usuário pode pertencer a múltiplos tenants)
CREATE TABLE tenant_members (
    user_id    UUID REFERENCES users(id)      ON DELETE CASCADE,
    tenant_id  UUID REFERENCES tenants(id)    ON DELETE CASCADE,
    role_id    UUID REFERENCES user_roles(id) ON DELETE SET NULL,
    is_owner   BOOLEAN   NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,  -- soft delete (remover membro sem perder histórico)
    PRIMARY KEY (user_id, tenant_id)
);
```

---

### 4.7 Tabelas de Dados do Tenant (Isolamento Lógico por tenant_id)

```sql
-- Produtos (tenant_id obrigatório)
CREATE TABLE products (
    id          UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID           NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(255)   NOT NULL,
    description TEXT,
    price       DECIMAL(10,2)  NOT NULL DEFAULT 0,
    sku         VARCHAR(100),
    stock       INTEGER        NOT NULL DEFAULT 0,
    is_active   BOOLEAN        NOT NULL DEFAULT true,
    image_url   TEXT,
    created_at  TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP      NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, sku)
);

-- Serviços (tenant_id obrigatório)
CREATE TABLE services (
    id          UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID           NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(255)   NOT NULL,
    description TEXT,
    price       DECIMAL(10,2)  NOT NULL DEFAULT 0,
    duration    INTEGER,       -- Duração em minutos (opcional)
    is_active   BOOLEAN        NOT NULL DEFAULT true,
    image_url   TEXT,
    created_at  TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP      NOT NULL DEFAULT NOW()
);

-- Configurações do tenant (key-value JSONB por categoria)
CREATE TABLE settings (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id  UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    category   VARCHAR(100) NOT NULL,  -- ex: 'interface', 'notifications', 'payments'
    data       JSONB        NOT NULL DEFAULT '{}',
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, category)
);
```

---

### 4.8 Tabelas de Usuários/Clientes do App do Tenant

```sql
-- Clientes/usuários finais (cadastrados pelo próprio tenant ou pelo app público)
CREATE TABLE tenant_app_users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL, 
    email         VARCHAR(255) NOT NULL,
    hash_pass     VARCHAR(255) NOT NULL,
    status        user_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP,   -- soft delete
    UNIQUE (tenant_id, email)  -- Email único por tenant (não global)
);

CREATE TABLE tenant_app_user_profiles (
    app_user_id UUID         PRIMARY KEY REFERENCES tenant_app_users(id) ON DELETE CASCADE,
    full_name   VARCHAR(255),
    phone       VARCHAR(30),
    document    VARCHAR(30),  -- CPF/CNPJ
    birth_date  DATE,
    avatar_url  TEXT,
    address     JSONB        NOT NULL DEFAULT '{}',
    metadata    JSONB        NOT NULL DEFAULT '{}',  -- Dados customizados pelo tenant
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);
```

---

### 4.9 Tabela de Imagens (Upload Centralizado)

```sql
CREATE TABLE images (
    id             UUID             PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id      UUID             NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    original_name  VARCHAR(500)     NOT NULL,
    storage_path   TEXT             NOT NULL,
    public_url     TEXT             NOT NULL,
    file_size      BIGINT           NOT NULL DEFAULT 0,
    mime_type      VARCHAR(100)     NOT NULL,
    width          INTEGER,
    height         INTEGER,
    provider       storage_provider NOT NULL DEFAULT 'local',
    entity_type    VARCHAR(50),     -- 'product', 'service', 'profile', etc.
    entity_id      UUID,            -- ID da entidade relacionada
    uploaded_by    UUID             REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMP        NOT NULL DEFAULT NOW()
);
```

---

### 4.10 Índices

```sql
-- Tenants
CREATE INDEX idx_tenants_url_code    ON tenants(url_code);
CREATE INDEX idx_tenants_subdomain   ON tenants(subdomain);
CREATE INDEX idx_tenants_status      ON tenants(status);
CREATE INDEX idx_tenants_deleted_at  ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;

-- Tenant plans
CREATE INDEX idx_tenant_plans_tenant_id ON tenant_plans(tenant_id);
CREATE INDEX idx_tenant_plans_plan_id   ON tenant_plans(plan_id);
-- (o índice UNIQUE parcial idx_tenant_plans_active já foi criado na seção 4.5)

-- Promotions
CREATE INDEX idx_promotions_is_active ON promotions(is_active);

-- Users
CREATE INDEX idx_users_email       ON users(email);
CREATE INDEX idx_users_status      ON users(status);
CREATE INDEX idx_users_deleted_at  ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Tenant members
CREATE INDEX idx_tenant_members_user_id    ON tenant_members(user_id)   WHERE deleted_at IS NULL;
CREATE INDEX idx_tenant_members_tenant_id  ON tenant_members(tenant_id) WHERE deleted_at IS NULL;

-- Dados dos tenants (CRÍTICO: todos devem ter índice em tenant_id)
CREATE INDEX idx_products_tenant_id  ON products(tenant_id);
CREATE INDEX idx_services_tenant_id  ON services(tenant_id);
CREATE INDEX idx_settings_tenant_id  ON settings(tenant_id);
CREATE INDEX idx_images_tenant_id    ON images(tenant_id);

-- App users
CREATE INDEX idx_tenant_app_users_tenant_id ON tenant_app_users(tenant_id);
CREATE INDEX idx_tenant_app_users_email     ON tenant_app_users(tenant_id, email)
    WHERE deleted_at IS NULL;

-- System admin
CREATE INDEX idx_system_admin_users_email      ON system_admin_users(email);
CREATE INDEX idx_system_admin_users_status     ON system_admin_users(status);
CREATE INDEX idx_system_admin_users_deleted_at ON system_admin_users(deleted_at)
    WHERE deleted_at IS NOT NULL;
```

---

## 5. Estrutura do Projeto

```
project-root/
├── cmd/
│   ├── tenant-api/      # Serviço Tenant Backoffice (porta 8080)
│   │   └── main.go
│   ├── admin-api/       # Serviço Admin SaaS (porta 8081)
│   │   └── main.go
│   └── app-api/         # Serviço App Users/Clientes (porta 8082)
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go         # Lê variáveis de ambiente
│   ├── database/
│   │   └── postgres.go       # Pool único pgxpool.Pool
│   ├── cache/
│   │   └── redis.go          # Wrapper Redis
│   ├── storage/
│   │   ├── interface.go      # interface Provider
│   │   ├── factory.go        # NewProvider(config)
│   │   ├── local.go
│   │   ├── s3.go
│   │   └── r2.go
│   ├── middleware/
│   │   ├── admin_auth.go     # JWT para system_admin_users
│   │   ├── user_auth.go      # JWT para users (backoffice)
│   │   ├── app_auth.go       # JWT para tenant_app_users
│   │   └── tenant.go         # Resolve tenant_id via url_code, injeta no ctx
│   ├── models/
│   │   ├── admin/            # Structs para system_admin_*
│   │   ├── tenant/           # Structs para tenants, plans, features
│   │   ├── user/             # Structs para users, tenant_members
│   │   ├── app/              # Structs para tenant_app_users
│   │   └── shared/           # Enums, paginação, resposta padrão
│   ├── repository/
│   │   ├── admin/
│   │   ├── tenant/
│   │   ├── user/
│   │   └── app/
│   ├── services/
│   │   ├── admin/
│   │   ├── tenant/
│   │   └── app/
│   ├── handlers/
│   │   ├── admin/
│   │   ├── tenant/
│   │   └── app/
│   └── utils/
│       ├── auth.go           # JWT generate/validate, bcrypt
│       ├── pagination.go     # Paginação padrão
│       └── slugger.go        # Geração de url_code/slug
├── migrations/
│   └── 001_initial_schema.up.sql
│   └── 001_initial_schema.down.sql
├── scripts/
│   └── makefiles/
│       ├── admin-tests.mk
│       ├── tenant-tests.mk
│       ├── user-tests.mk
│       ├── app-tests.mk
│       ├── product-tests.mk
│       ├── service-tests.mk
│       └── setting-tests.mk
├── uploads/                   # Upload local (montado via volume Docker)
├── docker-compose.yml
├── Makefile
├── go.mod
└── .env.example
```

---

## 6. Variáveis de Ambiente (`.env`)

```env
# Database
DATABASE_URL=postgres://saasuser:saaspassword@localhost:5432/saasdb?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=supersecretkey
JWT_EXPIRY_HOURS=24

# Storage (local | s3 | r2)
STORAGE_PROVIDER=local
STORAGE_LOCAL_PATH=./uploads
STORAGE_BASE_URL=http://localhost:8080/uploads

# S3 (se STORAGE_PROVIDER=s3)
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_REGION=us-east-1
AWS_BUCKET=

# R2 (se STORAGE_PROVIDER=r2)
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET=
R2_PUBLIC_URL=

# Ports
TENANT_API_PORT=8080
ADMIN_API_PORT=8081
APP_API_PORT=8082
```

---

## 7. Docker Compose

```yaml
version: '3.9'
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: saasdb
      POSTGRES_USER: saasuser
      POSTGRES_PASSWORD: saaspassword
    ports: ["5432:5432"]
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  tenant-api:
    build:
      context: .
      dockerfile: Dockerfile.tenant-api
    ports: ["8080:8080"]
    env_file: .env
    volumes:
      - ./uploads:/app/uploads
    depends_on: [postgres, redis]

  admin-api:
    build:
      context: .
      dockerfile: Dockerfile.admin-api
    ports: ["8081:8081"]
    env_file: .env
    depends_on: [postgres, redis]

  app-api:
    build:
      context: .
      dockerfile: Dockerfile.app-api
    ports: ["8082:8082"]
    env_file: .env
    depends_on: [postgres, redis]

volumes:
  pgdata:
```

---

## 8. Endpoints da API

### Convenções
- Todos os retornos são JSON.
- Erros seguem o padrão: `{"error": "mensagem"}` ou `{"errors": {...}}` para validação.
- Paginação: `?page=1&page_size=20` → resposta: `{"data": [...], "total": N, "page": 1, "page_size": 20}`.
- Autenticação: `Authorization: Bearer <jwt_token>`.
- O `tenant_id` **NUNCA** é enviado pelo client; é resolvido pelo middleware via `:url_code`.

---

### 8.1 Admin API — `http://localhost:8081`

#### Auth

| Método | Endpoint                     | Auth | Descrição                          |
|--------|------------------------------|------|------------------------------------|
| POST   | `/api/v1/admin/auth/login`   | —    | Login do administrador do SaaS     |
| POST   | `/api/v1/admin/auth/logout`  | ✅   | Invalida token (Redis blacklist)   |
| GET    | `/api/v1/admin/auth/me`      | ✅   | Retorna dados do admin logado      |
| PUT    | `/api/v1/admin/auth/password`| ✅   | Alterar senha do admin logado      |

**POST /api/v1/admin/auth/login**
```json
// Request
{ "email": "admin@saas.com", "password": "senha123" }

// Response 200
{
  "token": "eyJ...",
  "admin": { "id": "uuid", "email": "admin@saas.com", "profile": { "full_name": "Nome" } }
}
```

---

#### Gerenciamento de Admins (system_admin_users)

| Método | Endpoint                         | Auth | Permissão             | Descrição             |
|--------|----------------------------------|------|-----------------------|-----------------------|
| GET    | `/api/v1/admin/sys-users`        | ✅   | `manage_sys_users`    | Listar admins         |
| POST   | `/api/v1/admin/sys-users`        | ✅   | `manage_sys_users`    | Criar admin           |
| GET    | `/api/v1/admin/sys-users/:id`    | ✅   | `manage_sys_users`    | Detalhar admin        |
| PUT    | `/api/v1/admin/sys-users/:id`    | ✅   | `manage_sys_users`    | Atualizar admin       |
| DELETE | `/api/v1/admin/sys-users/:id`    | ✅   | `manage_sys_users`    | Desativar admin       |
| GET    | `/api/v1/admin/sys-users/:id/profile` | ✅ | `manage_sys_users` | Ver perfil         |
| PUT    | `/api/v1/admin/sys-users/:id/profile` | ✅ | `manage_sys_users` | Atualizar perfil   |
| GET    | `/api/v1/admin/sys-users/profile`     | ✅ | —                  | Meu perfil         |
| PUT    | `/api/v1/admin/sys-users/profile`     | ✅ | —                  | Atualizar meu perfil |

---

#### Gerenciamento de Roles/Permissões Admin

| Método | Endpoint                                            | Auth | Descrição                    |
|--------|-----------------------------------------------------|------|------------------------------|
| GET    | `/api/v1/admin/roles`                               | ✅   | Listar roles de admin        |
| POST   | `/api/v1/admin/roles`                               | ✅   | Criar role admin             |
| GET    | `/api/v1/admin/roles/:id`                           | ✅   | Detalhar role                |
| PUT    | `/api/v1/admin/roles/:id`                           | ✅   | Atualizar role               |
| DELETE | `/api/v1/admin/roles/:id`                           | ✅   | Deletar role                 |
| POST   | `/api/v1/admin/sys-users/:id/roles`                 | ✅   | Atribuir role a admin        |
| DELETE | `/api/v1/admin/sys-users/:id/roles/:role_id`        | ✅   | Remover role de admin        |
| GET    | `/api/v1/admin/permissions`                         | ✅   | Listar todas as permissões   |

---

#### Gerenciamento de Tenants

| Método | Endpoint                              | Auth | Permissão        | Descrição                        |
|--------|---------------------------------------|------|------------------|----------------------------------|
| GET    | `/api/v1/admin/tenants`               | ✅   | `view_tenants`   | Listar tenants (paginado)        |
| POST   | `/api/v1/admin/tenants`               | ✅   | `create_tenant`  | Criar tenant manualmente         |
| GET    | `/api/v1/admin/tenants/:id`           | ✅   | `view_tenants`   | Detalhar tenant                  |
| PUT    | `/api/v1/admin/tenants/:id`           | ✅   | `update_tenant`  | Atualizar tenant                 |
| DELETE | `/api/v1/admin/tenants/:id`           | ✅   | `delete_tenant`  | Cancelar/deletar tenant          |
| PUT    | `/api/v1/admin/tenants/:id/status`    | ✅   | `update_tenant`  | Mudar status (active/suspended)  |
| PUT    | `/api/v1/admin/tenants/:id/plan`          | ✅   | `manage_plans`   | Mudar plano do tenant (cria novo tenant_plan) |
| GET    | `/api/v1/admin/tenants/:id/plan-history`  | ✅   | `view_tenants`   | Histórico de planos contratados  |
| GET    | `/api/v1/admin/tenants/:id/members`       | ✅   | `view_tenants`   | Listar membros do tenant         |

**POST /api/v1/admin/tenants**
```json
// Request
{
  "name": "Empresa Teste",
  "url_code": "minha-empresa",
  "subdomain": "minha-empresa",
  "is_company": true,
  "company_name": "Empresa Ltda",
  "plan_id": "uuid-do-plano",
  "billing_cycle": "monthly",
  "promotion_id": "uuid-da-promo",   // opcional
  "owner_email": "dono@empresa.com", // opcional: cria/vincula usuário como owner
  "owner_full_name": "João Silva",
  "owner_password": "senha"
}
```

**PUT /api/v1/admin/tenants/:id/plan**
```json
// Request — desativa o tenant_plan ativo e cria um novo
{
  "plan_id": "uuid-novo-plano",
  "billing_cycle": "annual",
  "promotion_id": "uuid-promo",  // opcional
  "reason": "Upgrade solicitado pelo cliente" // opcional, para auditoria
}
```

---

#### Gerenciamento de Promoções

| Método | Endpoint                              | Auth | Permissão      | Descrição                          |
|--------|---------------------------------------|------|----------------|------------------------------------|
| GET    | `/api/v1/admin/promotions`            | ✅   | `manage_plans` | Listar promoções                   |
| POST   | `/api/v1/admin/promotions`            | ✅   | `manage_plans` | Criar promoção                     |
| GET    | `/api/v1/admin/promotions/:id`        | ✅   | `manage_plans` | Detalhar promoção                  |
| PUT    | `/api/v1/admin/promotions/:id`        | ✅   | `manage_plans` | Atualizar promoção                 |
| DELETE | `/api/v1/admin/promotions/:id`        | ✅   | `manage_plans` | Desativar promoção (soft)          |

**POST /api/v1/admin/promotions**
```json
// Request
{
  "name": "Lançamento 50% off",
  "description": "50% de desconto nos 3 primeiros meses",
  "discount_type": "percent",
  "discount_value": 50.00,
  "duration_months": 3,
  "valid_from": "2026-03-01T00:00:00Z",
  "valid_until": "2026-06-30T23:59:59Z"  // null = sem prazo para novos cadastros
}

// Response 201
{
  "id": "uuid",
  "name": "Lançamento 50% off",
  "discount_type": "percent",
  "discount_value": 50.00,
  "duration_months": 3,
  "example": "Plano R$ 99,90 → R$ 49,95 por 3 meses, depois R$ 99,90"
}
```

---

#### Gerenciamento de Planos e Features

| Método | Endpoint                                    | Auth | Descrição                  |
|--------|---------------------------------------------|------|----------------------------|
| GET    | `/api/v1/admin/plans`                       | ✅   | Listar planos              |
| POST   | `/api/v1/admin/plans`                       | ✅   | Criar plano                |
| GET    | `/api/v1/admin/plans/:id`                   | ✅   | Detalhar plano             |
| PUT    | `/api/v1/admin/plans/:id`                   | ✅   | Atualizar plano            |
| DELETE | `/api/v1/admin/plans/:id`                   | ✅   | Deletar plano              |
| POST   | `/api/v1/admin/plans/:id/features`          | ✅   | Adicionar feature ao plano |
| DELETE | `/api/v1/admin/plans/:id/features/:feat_id` | ✅   | Remover feature do plano   |
| GET    | `/api/v1/admin/features`                    | ✅   | Listar features            |
| POST   | `/api/v1/admin/features`                    | ✅   | Criar feature              |
| GET    | `/api/v1/admin/features/:id`                | ✅   | Detalhar feature           |
| PUT    | `/api/v1/admin/features/:id`                | ✅   | Atualizar feature          |
| DELETE | `/api/v1/admin/features/:id`                | ✅   | Deletar feature            |

---

### 8.2 Tenant API — `http://localhost:8080`

#### Subscription (Público)

| Método | Endpoint                        | Auth | Descrição                                      |
|--------|---------------------------------|------|------------------------------------------------|
| POST   | `/api/v1/subscription`          | —    | Cadastro público: cria tenant + owner          |
| GET    | `/api/v1/plans`                 | —    | Listar planos disponíveis (para landing page)  |

**POST /api/v1/subscription**
```json
// Request
{
  "plan_id": "uuid",
  "billing_cycle": "monthly",
  "promotion_id": "uuid-da-promo",  // opcional — validado contra valid_from/valid_until
  "name": "Minha Loja",
  "url_code": "minha-loja",
  "subdomain": "minha-loja",
  "is_company": false,
  "company_name": "",
  "full_name": "Maria Silva",
  "email": "maria@minha-loja.com",
  "password": "senha123"
}

// Response 201
{
  "tenant": {
    "id": "uuid",
    "name": "Minha Loja",
    "url_code": "minha-loja",
    "status": "active"
  },
  "subscription": {
    "plan": "Premium",
    "billing_cycle": "monthly",
    "contracted_price": 99.90,
    "promo_price": 49.95,
    "promo_expires_at": "2026-06-01T00:00:00Z",
    "promotion": "Lançamento 50% off — 3 meses"
  },
  "token": "eyJ...",
  "user": { "id": "uuid", "email": "maria@minha-loja.com" }
}
```

> **Nota:** Ao criar a subscription, a API calcula e persiste `promo_price` e `promo_expires_at` com base na `promotion` selecionada. O `contracted_price` recebe o valor de tabela do plano (`plan.price`), e o `promo_price` é calculado (`base_price - desconto`).

---

#### Auth (Backoffice Users)

| Método | Endpoint                           | Auth | Descrição                                   |
|--------|------------------------------------|------|---------------------------------------------|
| POST   | `/api/v1/auth/login`               | —    | Login do usuário backoffice                 |
| POST   | `/api/v1/auth/logout`              | ✅   | Invalida token                              |
| GET    | `/api/v1/auth/me`                  | ✅   | Dados do usuário + todos os seus tenants    |
| POST   | `/api/v1/auth/switch/:url_code`    | ✅   | Trocar tenant ativo, retorna novo JWT       |

**POST /api/v1/auth/login**
```json
// Request
{ "email": "maria@minha-loja.com", "password": "senha123" }

// Response 200
{
  "token": "eyJ...",  // JWT com tenant_id do last_tenant_url_code
  "user": {
    "id": "uuid",
    "email": "maria@minha-loja.com",
    "profile": { "full_name": "Maria Silva" }
  },
  "current_tenant": {
    "id": "uuid",
    "url_code": "minha-loja",
    "company_name": "Minha Loja",
    "features": ["products", "services"],
    "permissions": ["prod_c", "prod_r", "prod_u", "prod_d"]
  },
  "tenants": [...]  // Todos os tenants que o usuário pertence
}
```

---

#### Config do Tenant (Frontend Bridge)

| Método | Endpoint                              | Auth | Descrição                                              |
|--------|---------------------------------------|------|--------------------------------------------------------|
| GET    | `/api/v1/:url_code/config`            | ✅   | Features e permissions do usuário neste tenant         |

```json
// Response 200
{
  "tenant": {
    "id": "uuid",
    "name": "Minha Loja",
    "url_code": "minha-loja",
    "company_name": "Empresa Ltda"
  },
  "features": ["products", "services"],
  "permissions": ["prod_c", "prod_r", "serv_c", "serv_r"],
  "plan": {
    "name": "Premium",
    "max_users": 5,
    "current_users": 2,
    "available_slots": 3,
    "is_multilang": true,
    "billing_cycle": "monthly",
    "contracted_price": 99.90,
    "active_price": 49.95,          // promo_price se vigente, senão contracted_price
    "promo_expires_at": "2026-06-01T00:00:00Z",  // null se sem promoção ativa
    "price_updated_at": "2026-03-01T00:00:00Z"
  }
}
```

> O frontend usa `plan.is_multilang` para exibir/ocultar seletor de idiomas. Usa `plan.max_users` vs `plan.current_users` para desabilitar proativamente o botão de "Adicionar Membro" sem precisar chamar `can-add`.

---

#### Profile do Usuário Backoffice

| Método | Endpoint                              | Auth | Descrição                  |
|--------|---------------------------------------|------|----------------------------|
| GET    | `/api/v1/profile`                     | ✅   | Meu perfil                 |
| PUT    | `/api/v1/profile`                     | ✅   | Atualizar meu perfil       |
| PUT    | `/api/v1/profile/password`            | ✅   | Alterar minha senha        |
| POST   | `/api/v1/profile/avatar`              | ✅   | Upload avatar (multipart)  |

---

#### Gerenciamento do Tenant (dentro do backoffice)

| Método | Endpoint                              | Auth | Permissão    | Descrição                           |
|--------|---------------------------------------|------|--------------|-------------------------------------|
| GET    | `/api/v1/:url_code/tenant`            | ✅   | —            | Ver dados do tenant                 |
| PUT    | `/api/v1/:url_code/tenant/profile`    | ✅   | `setg_m`     | Atualizar perfil do tenant          |
| POST   | `/api/v1/:url_code/tenant/logo`       | ✅   | `setg_m`     | Upload logo do tenant               |

---

#### Membros do Tenant (Usuários Backoffice)

| Método | Endpoint                                       | Auth | Permissão | Descrição                                          |
|--------|------------------------------------------------|------|-----------|----------------------------------------------------|
| GET    | `/api/v1/:url_code/members`                    | ✅   | `user_m`  | Listar membros                                     |
| GET    | `/api/v1/:url_code/members/can-add`            | ✅   | `user_m`  | Verifica se pode adicionar membro (pré-tela)       |
| POST   | `/api/v1/:url_code/members`                    | ✅   | `user_m`  | Convidar/criar membro                              |
| GET    | `/api/v1/:url_code/members/:user_id`           | ✅   | `user_m`  | Ver membro                                         |
| PUT    | `/api/v1/:url_code/members/:user_id/role`      | ✅   | `user_m`  | Alterar role do membro                             |
| DELETE | `/api/v1/:url_code/members/:user_id`           | ✅   | `user_m`  | Remover membro do tenant                           |

**GET /api/v1/:url_code/members/can-add**

> Chamado pelo frontend **antes de exibir a tela de cadastro** de membro. A API compara o total atual de usuários ativos do tenant com o `max_users` do plano contratado (via `tenant_plans`). Retorna 200 com `can_add: true/false` e dados de contexto para o frontend exibir mensagem adequada.

```json
// Response 200 — dentro do limite
{
  "can_add": true,
  "current_users": 2,
  "max_users": 5,
  "available_slots": 3
}

// Response 200 — limite atingido
{
  "can_add": false,
  "current_users": 5,
  "max_users": 5,
  "available_slots": 0,
  "reason": "user_limit_reached",
  "upgrade_hint": "Faça upgrade do seu plano para adicionar mais colaboradores."
}
```

> **Implementação:** O `POST /members` também valida o limite como segunda barreira (never trust the client), retornando `HTTP 422` com `{"error": "user_limit_reached"}` se o limite for ultrapassado diretamente.

---

#### Roles e Permissões do Tenant

| Método | Endpoint                                              | Auth | Permissão | Descrição                        |
|--------|-------------------------------------------------------|------|-----------|----------------------------------|
| GET    | `/api/v1/:url_code/roles`                             | ✅   | `user_m`  | Listar roles do tenant           |
| POST   | `/api/v1/:url_code/roles`                             | ✅   | `user_m`  | Criar role customizada           |
| GET    | `/api/v1/:url_code/roles/:id`                         | ✅   | `user_m`  | Detalhar role                    |
| PUT    | `/api/v1/:url_code/roles/:id`                         | ✅   | `user_m`  | Atualizar role                   |
| DELETE | `/api/v1/:url_code/roles/:id`                         | ✅   | `user_m`  | Deletar role                     |
| GET    | `/api/v1/:url_code/roles/:id/permissions`             | ✅   | `user_m`  | Listar permissões da role        |
| POST   | `/api/v1/:url_code/roles/:id/permissions`             | ✅   | `user_m`  | Atribuir permissão à role        |
| DELETE | `/api/v1/:url_code/roles/:id/permissions/:perm_id`    | ✅   | `user_m`  | Remover permissão da role        |

---

#### Produtos

| Método | Endpoint                                 | Auth | Feature    | Permissão | Descrição              |
|--------|------------------------------------------|------|------------|-----------|------------------------|
| GET    | `/api/v1/:url_code/products`             | ✅   | `products` | `prod_r`  | Listar (paginado)      |
| POST   | `/api/v1/:url_code/products`             | ✅   | `products` | `prod_c`  | Criar produto          |
| GET    | `/api/v1/:url_code/products/:id`         | ✅   | `products` | `prod_r`  | Detalhar              |
| PUT    | `/api/v1/:url_code/products/:id`         | ✅   | `products` | `prod_u`  | Atualizar             |
| DELETE | `/api/v1/:url_code/products/:id`         | ✅   | `products` | `prod_d`  | Deletar (soft delete) |
| POST   | `/api/v1/:url_code/products/:id/image`   | ✅   | `products` | `prod_u`  | Upload imagem         |

---

#### Serviços

| Método | Endpoint                                 | Auth | Feature    | Permissão | Descrição              |
|--------|------------------------------------------|------|------------|-----------|------------------------|
| GET    | `/api/v1/:url_code/services`             | ✅   | `services` | `serv_r`  | Listar (paginado)      |
| POST   | `/api/v1/:url_code/services`             | ✅   | `services` | `serv_c`  | Criar serviço         |
| GET    | `/api/v1/:url_code/services/:id`         | ✅   | `services` | `serv_r`  | Detalhar             |
| PUT    | `/api/v1/:url_code/services/:id`         | ✅   | `services` | `serv_u`  | Atualizar            |
| DELETE | `/api/v1/:url_code/services/:id`         | ✅   | `services` | `serv_d`  | Deletar (soft delete)|
| POST   | `/api/v1/:url_code/services/:id/image`   | ✅   | `services` | `serv_u`  | Upload imagem        |

---

#### Configurações do Tenant

| Método | Endpoint                                        | Auth | Permissão | Descrição                          |
|--------|-------------------------------------------------|------|-----------|------------------------------------|
| GET    | `/api/v1/:url_code/settings`                    | ✅   | `setg_m`  | Listar todas as categorias         |
| GET    | `/api/v1/:url_code/settings/:category`          | ✅   | `setg_m`  | Ler configuração de uma categoria  |
| PUT    | `/api/v1/:url_code/settings/:category`          | ✅   | `setg_m`  | Atualizar configuração             |

---

#### Imagens (Upload)

| Método | Endpoint                             | Auth | Descrição                               |
|--------|--------------------------------------|------|-----------------------------------------|
| POST   | `/api/v1/:url_code/images`           | ✅   | Upload de imagem (retorna URLs)         |
| GET    | `/api/v1/:url_code/images`           | ✅   | Listar imagens do tenant (paginado)     |
| DELETE | `/api/v1/:url_code/images/:id`       | ✅   | Deletar imagem                          |

---

### 8.3 App API — `http://localhost:8082`

> API pública voltada para os usuários finais/clientes dos tenants.  
> O `tenant_id` é resolvido via `:url_code` na URL (mesmo middleware).

#### Auth App Users (Público)

| Método | Endpoint                                   | Auth | Descrição                              |
|--------|--------------------------------------------|------|----------------------------------------|
| POST   | `/api/v1/:url_code/auth/register`          | —    | Registro público de cliente            |
| POST   | `/api/v1/:url_code/auth/login`             | —    | Login do cliente                       |
| POST   | `/api/v1/:url_code/auth/logout`            | ✅   | Logout                                 |
| GET    | `/api/v1/:url_code/auth/me`                | ✅   | Dados do cliente logado                |
| POST   | `/api/v1/:url_code/auth/forgot-password`   | —    | Solicitar reset de senha               |
| POST   | `/api/v1/:url_code/auth/reset-password`    | —    | Resetar senha com token                |

**POST /api/v1/:url_code/auth/register**
```json
// Request
{
  "email": "cliente@exemplo.com",
  "password": "senha123",
  "full_name": "Carlos Santos",
  "phone": "+5511999999999"
}

// Response 201
{
  "token": "eyJ...",
  "user": { "id": "uuid", "email": "cliente@exemplo.com" }
}
```

---

#### Perfil do App User

| Método | Endpoint                                       | Auth | Descrição                  |
|--------|------------------------------------------------|------|----------------------------|
| GET    | `/api/v1/:url_code/profile`                    | ✅   | Ver meu perfil             |
| PUT    | `/api/v1/:url_code/profile`                    | ✅   | Atualizar meu perfil       |
| PUT    | `/api/v1/:url_code/profile/password`           | ✅   | Alterar senha              |
| POST   | `/api/v1/:url_code/profile/avatar`             | ✅   | Upload avatar              |

---

#### Catálogo Público (sem auth opcional)

| Método | Endpoint                                       | Auth | Descrição                          |
|--------|------------------------------------------------|------|------------------------------------|
| GET    | `/api/v1/:url_code/catalog/products`           | —    | Listar produtos ativos do tenant   |
| GET    | `/api/v1/:url_code/catalog/products/:id`       | —    | Detalhar produto                   |
| GET    | `/api/v1/:url_code/catalog/services`           | —    | Listar serviços ativos do tenant   |
| GET    | `/api/v1/:url_code/catalog/services/:id`       | —    | Detalhar serviço                   |

---

#### Gerenciamento de App Users pelo Tenant (backoffice → app users)

> Endpoints no tenant-api para o dono do tenant gerenciar seus clientes.

| Método | Endpoint                                          | Auth | Permissão | Descrição                       |
|--------|---------------------------------------------------|------|-----------|---------------------------------|
| GET    | `/api/v1/:url_code/app-users`                     | ✅   | `user_m`  | Listar clientes do tenant       |
| GET    | `/api/v1/:url_code/app-users/:id`                 | ✅   | `user_m`  | Detalhar cliente                |
| PUT    | `/api/v1/:url_code/app-users/:id/status`          | ✅   | `user_m`  | Ativar/suspender cliente        |
| DELETE | `/api/v1/:url_code/app-users/:id`                 | ✅   | `user_m`  | Deletar cliente                 |

---

## 9. Padrões de Código

### 9.1 Middleware de Tenant
```go
// Resolução do tenant_id a partir do :url_code
// 1. Extrair url_code do parâmetro de rota
// 2. Buscar tenant_id no Redis: cache key = "tenant:urlcode:{url_code}"
// 3. Se cache miss: query no PostgreSQL → SET no Redis com TTL de 5 minutos
// 4. Injetar tenant_id no contexto: c.Set("tenant_id", tenantID)
// 5. Injetar features ativas: c.Set("features", []string{"products", "services"})
// 6. Verificar se tenant está "active"; se não, retornar 403
func TenantMiddleware(db *pgxpool.Pool, cache *redis.Client) gin.HandlerFunc
```

### 9.2 Middleware de Auth (3 variantes)
```go
// Para system_admin_users → chave JWT: "admin:token"
func AdminAuthMiddleware(jwtSecret string) gin.HandlerFunc

// Para users (backoffice) → chave JWT: "user:token", injeta user_id + tenant_id
func UserAuthMiddleware(jwtSecret string) gin.HandlerFunc

// Para tenant_app_users → chave JWT: "app:token", injeta app_user_id + tenant_id
func AppAuthMiddleware(jwtSecret string) gin.HandlerFunc
```

### 9.3 Claims do JWT
```go
// Backoffice user
type UserClaims struct {
    UserID   string `json:"user_id"`
    TenantID string `json:"tenant_id"`
    jwt.RegisteredClaims
}

// App user
type AppUserClaims struct {
    AppUserID string `json:"app_user_id"`
    TenantID  string `json:"tenant_id"`
    jwt.RegisteredClaims
}

// Admin
type AdminClaims struct {
    AdminID string `json:"admin_id"`
    jwt.RegisteredClaims
}
```

### 9.4 Verificação de Feature nos Controllers
```go
func CreateProduct(c *gin.Context) {
    features := c.MustGet("features").([]string)
    if !slices.Contains(features, "products") {
        c.JSON(403, gin.H{"error": "feature_disabled"})
        return
    }
    tenantID := c.MustGet("tenant_id").(string)
    // ... lógica usando tenantID em todas as queries
}
```

### 9.5 Queries (SEMPRE filtrar por tenant_id)
```go
// CORRETO
const query = `SELECT * FROM products WHERE tenant_id = $1 AND id = $2`
row := db.QueryRow(ctx, query, tenantID, productID)

// ERRADO (vazamento entre tenants)
const query = `SELECT * FROM products WHERE id = $1`
```

### 9.6 Resposta Padrão de Paginação
```go
type PaginatedResponse struct {
    Data     interface{} `json:"data"`
    Total    int64       `json:"total"`
    Page     int         `json:"page"`
    PageSize int         `json:"page_size"`
}
```

---

## 10. Seed Data (Migração Inicial)

```sql
-- Roles do sistema admin
INSERT INTO system_admin_roles (name, slug) VALUES
    ('Super Admin', 'super_admin'),
    ('Admin',       'admin'),
    ('Support',     'support'),
    ('Viewer',      'viewer');

-- Permissões do sistema admin
INSERT INTO system_admin_permissions (name, slug) VALUES
    ('Manage Tenants',      'manage_tenants'),
    ('View Tenants',        'view_tenants'),
    ('Manage Plans',        'manage_plans'),
    ('Manage Features',     'manage_features'),
    ('Manage Sys Users',    'manage_sys_users'),
    ('View Analytics',      'view_analytics'),
    ('Manage Billing',      'manage_billing');

-- Admin padrão (email: admin@saas.com / senha: admin123)
-- Hash bcrypt custo 12 de 'admin123'
INSERT INTO system_admin_users (email, hash_pass) VALUES
    ('admin@saas.com', '$2a$12$...');

INSERT INTO system_admin_profiles (admin_user_id, full_name) 
    SELECT id, 'System Administrator' FROM system_admin_users WHERE email = 'admin@saas.com';

-- Features padrão
INSERT INTO features (id, title, slug, code) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Products', 'products', 'prod'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Services', 'services', 'serv');

-- Promoção de exemplo: 50% off nos 3 primeiros meses
INSERT INTO promotions (id, name, description, discount_type, discount_value, duration_months, valid_from, is_active) VALUES
    ('pppppppp-pppp-pppp-pppp-pppppppppppp',
     'Lançamento 50% off',
     '50% de desconto nos primeiros 3 meses',
     'percent', 50.00, 3,
     NOW(), true);

-- Planos padrão
-- max_users: 1=só owner, 3=owner+2 colaboradores, 5=owner+4, 10=owner+9
INSERT INTO plans (id, name, price, max_users, is_multilang) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Starter',    29.90,   1, false),
    ('22222222-2222-2222-2222-222222222222', 'Business',   59.90,   3, false),
    ('33333333-3333-3333-3333-333333333333', 'Premium',    99.90,   5, true),
    ('44444444-4444-4444-4444-444444444444', 'Enterprise', 199.90, 10, true);

-- Permissões de usuário (backoffice)
INSERT INTO user_permissions (name, slug, feature_id) VALUES
    ('Create Product', 'prod_c', 'aaaaaaaa-...'),
    ('Read Product',   'prod_r', 'aaaaaaaa-...'),
    ('Update Product', 'prod_u', 'aaaaaaaa-...'),
    ('Delete Product', 'prod_d', 'aaaaaaaa-...'),
    ('Create Service', 'serv_c', 'bbbbbbbb-...'),
    ('Read Service',   'serv_r', 'bbbbbbbb-...'),
    ('Update Service', 'serv_u', 'bbbbbbbb-...'),
    ('Delete Service', 'serv_d', 'bbbbbbbb-...'),
    ('Manage Users',   'user_m', NULL),
    ('Manage Settings','setg_m', NULL);

-- Roles globais de usuário (tenant_id NULL = templates copiados ao criar tenant)
INSERT INTO user_roles (tenant_id, name, slug) VALUES
    (NULL, 'Owner',  'owner'),
    (NULL, 'Admin',  'admin'),
    (NULL, 'Member', 'member');
```

---

## 11. Makefile — Targets Principais

```makefile
include scripts/makefiles/admin-tests.mk
include scripts/makefiles/tenant-tests.mk
include scripts/makefiles/user-tests.mk
include scripts/makefiles/app-tests.mk
include scripts/makefiles/product-tests.mk
include scripts/makefiles/service-tests.mk
include scripts/makefiles/setting-tests.mk

.PHONY: up down migrate logs build

up:
    docker compose up -d

down:
    docker compose down

build:
    docker compose build

migrate:
    docker compose exec postgres psql -U saasuser -d saasdb -f /migrations/001_initial_schema.up.sql

logs-admin:
    docker compose logs -f admin-api

logs-tenant:
    docker compose logs -f tenant-api

logs-app:
    docker compose logs -f app-api
```

---

## 12. scripts/makefiles/admin-tests.mk

```makefile
ADMIN_URL=http://localhost:8081
ADMIN_EMAIL=admin@saas.com
ADMIN_PASS=admin123

# ─── Helpers ───────────────────────────────────────────────
define get_admin_token
$(shell curl -s -X POST $(ADMIN_URL)/api/v1/admin/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"$(ADMIN_EMAIL)","password":"$(ADMIN_PASS)"}' \
    | grep -o '"token":"[^"]*' | cut -d'"' -f4)
endef

# ─── Auth ──────────────────────────────────────────────────
test-admin-login:
    @curl -s -X POST $(ADMIN_URL)/api/v1/admin/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(ADMIN_EMAIL)","password":"$(ADMIN_PASS)"}' | jq .
    @echo ""

test-admin-me:
    @TOKEN=$(call get_admin_token); \
    curl -s $(ADMIN_URL)/api/v1/admin/auth/me \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

# ─── Sys Users ─────────────────────────────────────────────
test-sysusers-list:
    @TOKEN=$(call get_admin_token); \
    curl -s $(ADMIN_URL)/api/v1/admin/sys-users \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-sysusers-create:
    @TOKEN=$(call get_admin_token); \
    curl -s -X POST $(ADMIN_URL)/api/v1/admin/sys-users \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $$TOKEN" \
        -d '{"email":"support@saas.com","password":"suporte123","full_name":"Suporte SaaS","role_slug":"support"}' | jq .
    @echo ""

# ─── Plans ─────────────────────────────────────────────────
test-plans-list:
    @TOKEN=$(call get_admin_token); \
    curl -s $(ADMIN_URL)/api/v1/admin/plans \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-plans-create:
    @TOKEN=$(call get_admin_token); \
    curl -s -X POST $(ADMIN_URL)/api/v1/admin/plans \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $$TOKEN" \
        -d '{"name":"Enterprise","description":"Plano completo","price":199.90,"feature_ids":["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"]}' | jq .
    @echo ""

# ─── Features ──────────────────────────────────────────────
test-features-list:
    @TOKEN=$(call get_admin_token); \
    curl -s $(ADMIN_URL)/api/v1/admin/features \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-features-create:
    @TOKEN=$(call get_admin_token); \
    curl -s -X POST $(ADMIN_URL)/api/v1/admin/features \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $$TOKEN" \
        -d '{"title":"Blog","slug":"blog","code":"blog","description":"Módulo de blog","is_active":true}' | jq .
    @echo ""

# ─── Tenants ───────────────────────────────────────────────
test-tenants-list:
    @TOKEN=$(call get_admin_token); \
    curl -s "$(ADMIN_URL)/api/v1/admin/tenants?page=1&page_size=10" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-tenants-create:
    @TOKEN=$(call get_admin_token); \
    curl -s -X POST $(ADMIN_URL)/api/v1/admin/tenants \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $$TOKEN" \
        -d '{"url_code":"empresa-teste","subdomain":"empresa-teste","plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","company_name":"Empresa Teste Ltda","is_company":true,"owner_email":"dono@empresa-teste.com","owner_full_name":"Dono Empresa","owner_password":"senha12345"}' | jq .
    @echo ""
```

---

## 13. scripts/makefiles/tenant-tests.mk

```makefile
TENANT_URL=http://localhost:8080

# ─── Subscription (Público) ────────────────────────────────
test-subscription:
    @echo "Criando novo tenant via subscription pública..."
    @curl -s -X POST $(TENANT_URL)/api/v1/subscription \
        -H "Content-Type: application/json" \
        -d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","name":"Minha Loja","url_code":"minha-loja","subdomain":"minha-loja","is_company":false,"full_name":"João Silva","email":"joao@minha-loja.com","password":"senha12345"}' | jq .
    @echo ""

test-subscription-with-promo:
    @echo "Criando tenant com promoção..."
    @curl -s -X POST $(TENANT_URL)/api/v1/subscription \
        -H "Content-Type: application/json" \
        -d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","promotion_id":"pppppppp-pppp-pppp-pppp-pppppppppppp","name":"Loja Promo","url_code":"loja-promo","subdomain":"loja-promo","is_company":false,"full_name":"Maria Promo","email":"maria@loja-promo.com","password":"senha12345"}' | jq .
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
        -d '{"plan_id":"33333333-3333-3333-3333-333333333333","billing_cycle":"monthly","name":"Nova Empresa","url_code":"nova-empresa","subdomain":"nova-empresa","is_company":false,"full_name":"Novo Usuario","email":"novo@empresa.com","password":"senha12345"}'); \
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
```

---

## 14. scripts/makefiles/product-tests.mk

```makefile
TENANT_URL=http://localhost:8080
USER_EMAIL=joao@minha-loja.com
USER_PASS=senha12345

define get_login
$(shell curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"$(USER_EMAIL)","password":"$(USER_PASS)"}')
endef

test-product-create:
    @LOGIN=$(call get_login); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $$TOKEN" \
        -d '{"name":"Notebook Dell","description":"Intel i7, 16GB RAM","price":3500.00,"sku":"NB-DELL-001","stock":10}' | jq .
    @echo ""

test-product-list:
    @LOGIN=$(call get_login); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/products?page=1&page_size=10" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-product-update:
    @LOGIN=$(call get_login); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/products/$(PRODUCT_ID)" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $$TOKEN" \
        -d '{"name":"Notebook Dell Atualizado","price":3200.00,"stock":15}' | jq .
    @echo ""

test-product-delete:
    @LOGIN=$(call get_login); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/products/$(PRODUCT_ID)" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-products-all:
    @echo "================================================"
    @echo "Teste CRUD completo de Produtos"
    @echo "================================================"
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(USER_EMAIL)","password":"$(USER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    echo "1. Criando produto..."; \
    PROD=$$(curl -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/products" \
        -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" \
        -d '{"name":"Notebook Dell","description":"Intel i7","price":3500.00,"sku":"NB-DELL-001","stock":10}'); \
    echo "$$PROD" | jq .; \
    PROD_ID=$$(echo "$$PROD" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4); \
    echo ""; echo "2. Listando produtos..."; \
    curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/products" -H "Authorization: Bearer $$TOKEN" | jq .; \
    echo ""; echo "3. Atualizando produto $$PROD_ID..."; \
    curl -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PROD_ID" \
        -H "Content-Type: application/json" -H "Authorization: Bearer $$TOKEN" \
        -d '{"price":3100.00}' | jq .; \
    echo ""; echo "4. Deletando produto $$PROD_ID..."; \
    curl -s -X DELETE "$(TENANT_URL)/api/v1/$$URL_CODE/products/$$PROD_ID" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""
```

---

## 15. scripts/makefiles/app-tests.mk

```makefile
APP_URL=http://localhost:8082
TENANT_URL=http://localhost:8080

test-app-register:
    @echo "Registrando cliente no app do tenant..."
    @curl -s -X POST $(APP_URL)/api/v1/minha-loja/auth/register \
        -H "Content-Type: application/json" \
        -d '{"email":"cliente@exemplo.com","password":"senha123","full_name":"Carlos Santos","phone":"+5511999999999"}' | jq .
    @echo ""

test-app-login:
    @curl -s -X POST $(APP_URL)/api/v1/minha-loja/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"cliente@exemplo.com","password":"senha123"}' | jq .
    @echo ""

test-app-catalog:
    @echo "Catálogo público de produtos..."
    @curl -s $(APP_URL)/api/v1/minha-loja/catalog/products | jq .
    @echo ""
    @echo "Catálogo público de serviços..."
    @curl -s $(APP_URL)/api/v1/minha-loja/catalog/services | jq .
    @echo ""

test-app-profile:
    @LOGIN=$$(curl -s -X POST $(APP_URL)/api/v1/minha-loja/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"cliente@exemplo.com","password":"senha123"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    curl -s $(APP_URL)/api/v1/minha-loja/profile \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-app-all:
    @echo "================================================"
    @echo "Teste E2E: App User (cliente do tenant)"
    @echo "================================================"
    @echo "1. Registro..."; \
    curl -s -X POST $(APP_URL)/api/v1/minha-loja/auth/register \
        -H "Content-Type: application/json" \
        -d '{"email":"cliente2@exemplo.com","password":"senha123","full_name":"Cliente Dois"}' | jq .; \
    echo "2. Login..."; \
    LOGIN=$$(curl -s -X POST $(APP_URL)/api/v1/minha-loja/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"cliente2@exemplo.com","password":"senha123"}'); \
    echo "$$LOGIN" | jq .; \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    echo "3. Perfil..."; \
    curl -s $(APP_URL)/api/v1/minha-loja/profile \
        -H "Authorization: Bearer $$TOKEN" | jq .; \
    echo "4. Catálogo..."; \
    curl -s "$(APP_URL)/api/v1/minha-loja/catalog/products" | jq .
    @echo ""
```

---

## 16. scripts/makefiles/setting-tests.mk

```makefile
TENANT_URL=http://localhost:8080
USER_EMAIL=joao@minha-loja.com
USER_PASS=senha12345

test-settings-list:
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(USER_EMAIL)","password":"$(USER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/settings" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-settings-get:
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(USER_EMAIL)","password":"$(USER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/settings/interface" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-settings-update:
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(USER_EMAIL)","password":"$(USER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s -X PUT "$(TENANT_URL)/api/v1/$$URL_CODE/settings/interface" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $$TOKEN" \
        -d '{"theme":"dark","language":"pt-BR","items_per_page":20}' | jq .
    @echo ""
```

---

## 17. scripts/makefiles/user-tests.mk

```makefile
TENANT_URL=http://localhost:8080
OWNER_EMAIL=joao@minha-loja.com
OWNER_PASS=senha12345

test-members-list:
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(OWNER_EMAIL)","password":"$(OWNER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/members" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-members-can-add:
    @echo "Verificando se pode adicionar membro (pré-tela de cadastro)..."
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(OWNER_EMAIL)","password":"$(OWNER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/members/can-add" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-members-invite:
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(OWNER_EMAIL)","password":"$(OWNER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s -X POST "$(TENANT_URL)/api/v1/$$URL_CODE/members" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $$TOKEN" \
        -d '{"email":"colaborador@minha-loja.com","full_name":"Colaborador","password":"senha12345","role_slug":"member"}' | jq .
    @echo ""

test-roles-list:
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(OWNER_EMAIL)","password":"$(OWNER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/roles" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""

test-app-users-list:
    @echo "Listando clientes do tenant (backoffice)..."
    @LOGIN=$$(curl -s -X POST $(TENANT_URL)/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"$(OWNER_EMAIL)","password":"$(OWNER_PASS)"}'); \
    TOKEN=$$(echo "$$LOGIN" | grep -o '"token":"[^"]*' | cut -d'"' -f4); \
    URL_CODE=$$(echo "$$LOGIN" | grep -o '"url_code":"[^"]*' | head -1 | cut -d'"' -f4); \
    curl -s "$(TENANT_URL)/api/v1/$$URL_CODE/app-users" \
        -H "Authorization: Bearer $$TOKEN" | jq .
    @echo ""
```

---

## 18. Anti-Padrões a Evitar

| ❌ Errado                                                                | ✅ Correto                                                                       |
|-------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| `SELECT * FROM products WHERE id = $1`                       | `SELECT * FROM products WHERE tenant_id = $1 AND id = $2`         |
| Criar pools de conexão por request                           | Pool único `pgxpool.Pool` compartilhado (conexão única ao DB)                             |
| Hardcodar `tenant_id` nas queries                            | Sempre receber `tenant_id` do contexto Gin (`c.MustGet`)                                  |
| Usar o mesmo JWT struct para admin e usuário                 | Tipos distintos de Claims para cada audiência                                             |
| Retornar objeto da tabela diretamente                        | Usar DTOs de resposta (nunca expor `hash_pass`)                                        |
| Confiar no `tenant_id` enviado pelo cliente                  | Resolver `tenant_id` exclusivamente via middleware                                         |
| `users` e `system_admin_users` compartilhando endpoint auth  | Endpoints e routers completamente separados por tipo de usuário                            |
| Só barrar limite de usuários no `POST /members`              | Checar **antes** via `GET /members/can-add` + barrar no POST também (double-check)        |
| Ignorar `is_multilang` no frontend                           | Ler a flag do `GET /config` e habilitar/desabilitar seleção de idioma conforme contrato   |
| `DELETE` físico em users, tenants ou members                 | Usar soft delete: `UPDATE ... SET deleted_at = NOW()`, nunca `DELETE FROM`                |
| Calcular preço de promoção no frontend                       | Persistir `promo_price` e `promo_expires_at` na tabela `tenant_plans` ao criar/alterar   |
| Consultar `plan_id` direto em `tenants`                      | Consultar `tenant_plans WHERE tenant_id = X AND is_active = true` para obter plano atual  |

---

## 19. Instruções para o Copilot (Resumo)

Ao construir este projeto em branco, siga esta ordem:

1. **Scaffold**: `go mod init`, criar estrutura de pastas.
2. **Configuração** (`internal/config/config.go`): Ler `.env` com `os.Getenv`.
3. **Database** (`internal/database/postgres.go`): `pgxpool.New` com pool único.
4. **Cache** (`internal/cache/redis.go`): `redis.NewClient` wrapper.
5. **Storage** (`internal/storage/`): Interface + implementações local/S3/R2.
6. **Migrations** (`migrations/`): Schema SQL completo com seed data.
7. **Models** (`internal/models/`): Structs Go para cada tabela + DTOs de request/response.
8. **Middleware** (`internal/middleware/`): `TenantMiddleware`, `AdminAuthMiddleware`, `UserAuthMiddleware`, `AppAuthMiddleware`.
9. **Repositories** (`internal/repository/`): Funções de acesso ao DB, sempre com `tenant_id`.
10. **Services** (`internal/services/`): Lógica de negócio, incluindo `tenant_plans` service (cálculo de promoção).
11. **Handlers** (`internal/handlers/`): Controllers Gin, verificam feature + permissão.
12. **Routers em cada `cmd/*/main.go`**: Registrar rotas com middlewares corretos.
13. **Docker Compose**: Postgres + Redis + 3 serviços.
14. **Makefile + scripts**: Targets de infraestrutura e testes via curl.

### Regras de Soft Delete (obrigatórias em todos os repositories)

```go
// SEMPRE adicionar ao WHERE em listagens
WHERE deleted_at IS NULL

// SEMPRE usar update ao invés de delete
UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1
// Nunca: DELETE FROM users WHERE id = $1

// Restaurar (quando suportado)
UPDATE users SET deleted_at = NULL, status = 'active', updated_at = NOW() WHERE id = $1
```

### Lógica de Preço Vigente (`tenant_plans`)

```go
// Para saber o preço a cobrar hoje:
func EffectivePrice(tp TenantPlan) decimal.Decimal {
    if tp.PromoExpiresAt != nil && tp.PromoExpiresAt.After(time.Now()) && tp.PromoPrice != nil {
        return *tp.PromoPrice
    }
    return tp.ContractedPrice
}

// Para obter o plano ativo de um tenant:
SELECT tp.*, p.name, p.max_users, p.is_multilang
FROM tenant_plans tp
JOIN plans p ON p.id = tp.plan_id
WHERE tp.tenant_id = $1 AND tp.is_active = true
LIMIT 1

// Para trocar de plano (transaction obrigatória):
// 1. UPDATE tenant_plans SET is_active = false, ended_at = NOW() WHERE tenant_id = $1 AND is_active = true
// 2. INSERT INTO tenant_plans (...) VALUES (...)
```
