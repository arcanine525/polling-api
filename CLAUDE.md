# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Polling system built as a microservice architecture with two Go services sharing a PostgreSQL database:

- **contact-api** (port 8080) — CSV upload and contact management (existing, working)
- **polling-api** (port 8081) — Polls, candidates, questions, answers, users (planned, see `SYSTEM_ARC.MD`)

Both services use Go/Gin/GORM and follow the same layered architecture: `handlers/ → services/ → database/` (GORM).

## Common Commands

### contact-api

```bash
cd contact-api

# Install dependencies
go mod tidy

# Run the server (starts on :8080)
go run main.go

# Build binary
go build -o contact-api .

# Regenerate Swagger docs (requires swag CLI)
swag init

# Run with PostgreSQL instead of SQLite
DB_DRIVER=postgres DB_DSN="host=localhost user=postgres password=postgres dbname=contacts port=5432 sslmode=disable" go run main.go
```

### Docker

```bash
cd contact-api
docker build -t contact-api .
docker run -p 8080:8080 contact-api
```

## Architecture

### Layered pattern (both services)

```
handlers/    HTTP handlers — parse request, call service, return JSON
services/    Business logic — validation, phone normalization, CSV parsing
database/    GORM init, auto-migration. Exports a global `database.DB` singleton
models/      GORM models, DTOs (create/update structs), paginated response types
config/      Env-based config loaded via godotenv with fallback defaults
middleware/  Auth middleware (Firebase token verification)
```

Handlers call service functions directly, passing the global `database.DB`. There is no dependency injection or interface-based abstraction — services accept `*gorm.DB` as the first parameter.

### Database

- contact-api defaults to **SQLite** (`contacts.db`), switchable to PostgreSQL via `DB_DRIVER` env var
- polling-api targets **PostgreSQL** only
- GORM `AutoMigrate` runs on startup — no separate migration tool
- Polling schema has 5 tables (`users`, `polls`, `candidates`, `poll_questions`, `answers`) with custom enum types (`user_role`, `poll_status`, `voice_poll_method`, `answer_type`, `answer_source`). Full schema in `init_db_schema.md`

### Authentication

Firebase Auth with middleware in each service (not a gateway). Users auto-provisioned on first request via firebase_uid upsert. RBAC roles: `admin`, `poller`, `viewer`. See `authentication_service_plan.md` for implementation details.



### Key conventions

- All list endpoints return a consistent paginated envelope: `{ items, total, page, size, pages }`
- Error responses: `{ "error": "message" }`
- All API routes prefixed with `/api/v1`
- Phone numbers normalized to E.164 format (Australian default: 9-digit → +61 prefix)
- Swagger annotations on all handler functions (swag format)
- GORM models use `BeforeCreate` hooks for UUID generation
- CSV uploads batch-insert contacts in groups of 1000 within a transaction


## Key Files

- `SYSTEM_ARC.MD` — Full system architecture, ER diagram, API endpoints, implementation order
- `init_db_schema.md` — Detailed polling database schema with all columns, constraints, and enums
- `authentication_service_plan.md` — Firebase Auth integration plan for Go and Python services

## Rules
- Always follow the layered architecture pattern when adding new features
- Use the existing code style and conventions for consistency
- Allways follow the tasks and update status in `tasks` folder, and update `SYSTEM_ARC.MD` when adding new endpoints or database tables