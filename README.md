# Go Backend (Production-Ready Boilerplate)

This repository is a production-ready Go backend template built with Gin, Supabase Postgres, Redis, Liquibase, Prometheus metrics, and a unified API response format. It is designed to be used as a boilerplate for new services or a CLI-driven project generator.

---

## Quick Start (Local)

Prerequisites:
- Go 1.25+
- Liquibase CLI
- Java 17+ (for Liquibase)
- `swag` CLI (for swagger generation)
- Optional: Redis (if `redis.enabled=true`)

One command setup and run:
```bash
make bootstrap
```

What it does:
- `make deps`
- `make db-driver` (downloads SQLite JDBC driver for Liquibase)
- `make db-migrate` (Liquibase migrations)
- `make swagger` (OpenAPI docs)
- `make run`

If you prefer step-by-step:
```bash
make deps
make db-driver
make db-migrate
make swagger
make run
```

---

## Configuration

Configuration uses Viper with a `config.yaml` file (optional) + environment overrides.

An `.env.example` is included and can be copied to `.env` for local development.

Example config:
```yaml
env: dev

server:
  address: ":8080"
  read_timeout: 5s
  write_timeout: 10s
  idle_timeout: 60s
  body_limit_bytes: 1048576
  shutdown_timeout: 10s

db:
  driver: postgres
  dsn: postgres://USER:PASSWORD@HOST:5432/DB?sslmode=require

redis:
  enabled: true
  addr: localhost:6379

rate_limit:
  default:
    rps: 10
    burst: 20
  use_redis: false
  redis_prefix: rate_limit
  routes:
    "GET /todos":
      rps: 15
      burst: 30
    "GET /todos/:id":
      rps: 15
      burst: 30
    "POST /todos":
      rps: 5
      burst: 10
    "PUT /todos/:id":
      rps: 5
      burst: 10

cors:
  enabled: false
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowed_headers: ["Authorization", "Content-Type", "X-Request-ID"]
  expose_headers: ["X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"]
  max_age_seconds: 600

auth:
  enabled: false
  api_key: ""

metrics:
  enabled: true
  path: /metrics
```

Environment variables follow Viper rules (e.g. `DB_DSN`, `RATE_LIMIT_DEFAULT_RPS`, `RATE_LIMIT_ROUTES__GET /todos__RPS`).

---

## Migrations (Liquibase)

Liquibase changelog:
- `liquibase/changelog.yaml`

Liquibase properties:
- `liquibase.properties`

Commands:
```bash
make db-driver
make db-migrate
make db-status
make db-rollback COUNT=1
```

Liquibase uses environment variables (via `liquibase.properties`):
- `DB_JDBC_URL` (example: `jdbc:postgresql://db.<project>.supabase.co:5432/postgres`)
- `DB_USERNAME`
- `DB_PASSWORD`

---

## API

Base URL: `http://localhost:8080`

Endpoints:
- `GET /health`
- `GET /ready`
- `GET /metrics`
- `GET /todos?limit=20&offset=0`
- `GET /todos/:id`
- `POST /todos`
- `PUT /todos/:id`

Response envelope:
```json
{
  "response": {},
  "status": 200,
  "message": "ok"
}
```

Rate-limit headers:
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`

---

## Swagger / OpenAPI

Generate docs:
```bash
make swagger
```

Swagger UI (dev only): `http://localhost:8080/swagger-ui/index.html`

---

## Docker

Build:
```bash
make docker-build
```

Run:
```bash
make docker-run
```

---

## Make Targets

Key targets:
- `make bootstrap` (one-shot setup + run)
- `make deps`
- `make db-driver`
- `make db-migrate`
- `make swagger`
- `make run`
- `make test`
- `make check`

---

## Using This As a Boilerplate (New Project)

Suggested steps:
1. Copy this repository to a new directory.
2. Rename module in `go.mod` to your new module path.
3. Replace all import paths to match the new module path.
   Example:
   ```bash
   rg -l "github.com/AshishBagdane/go-backend" | xargs sed -i '' 's#github.com/AshishBagdane/go-backend#github.com/your-org/your-service#g'
   ```
4. Update `cmd/server/main.go` metadata:
   - Swagger title, description, base path
5. Update `liquibase/changelog.yaml` with your schema.
6. Update `config.example.yaml` defaults and env requirements.
7. Replace `internal/api/handlers` with your own domain handlers.

---

## CLI Generator Plan (For Your Future Command)

If you plan to build a CLI that generates new projects:
- Template this repo as a scaffold.
- Parameterize:
  - module path
  - service name
  - default ports
  - database driver + DSN
  - rate limit defaults
- Apply string replacements across:
  - `go.mod`
  - imports in `cmd/` and `internal/`
  - swagger annotations

I can help build that CLI next if you want.
