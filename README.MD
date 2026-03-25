# Polling System

Microservice-style polling platform built with Go, featuring poll management, candidate tracking, question/answer workflows, contact management, and CSV bulk imports — all secured with Firebase Auth and role-based access control.

## Tech Stack

- **Go 1.26** / Gin / GORM
- **PostgreSQL**
- **Firebase Admin SDK** (authentication)
- **Swagger/OpenAPI** (swaggo)

## Project Structure

```
polling-system/
├── cmd/main.go              # Entry point
├── auth/                    # Authentication & user management
│   ├── handlers/
│   ├── models/
│   ├── services/
│   └── routes.go
├── polling/                 # Polls, candidates, questions, answers
│   ├── handlers/
│   ├── models/
│   ├── services/
│   └── routes.go
├── contact/                 # Contacts & CSV file uploads
│   ├── handlers/
│   ├── models/
│   ├── services/
│   └── routes.go
├── middleware/               # Auth, RBAC, user provisioning
├── config/                   # Env-based configuration
├── database/                 # GORM init & auto-migration
├── handlers/                 # Shared utilities (pagination)
├── docs/                     # Generated Swagger docs
└── tests/                    # Test suite
```

## Getting Started

### Prerequisites

- Go 1.26+
- PostgreSQL 12+
- Firebase project with service account credentials

### Setup

```bash
# Install dependencies
go mod tidy

# Configure environment
cp .env.example .env
# Edit .env with your database DSN, Firebase credentials, etc.

# Create database
createdb polling

# Run the server
go run cmd/main.go
```

The server starts on `http://localhost:8080` with auto-migration enabled.

### Docker

```bash
# Build
docker build -t polling-system .

# Run
docker run -p 8080:8080 \
  -e DB_DSN="postgresql://user:pass@host:5432/polling?sslmode=disable" \
  -e FIREBASE_CREDENTIALS_PATH="/app/serviceAccountKey.json" \
  -e FIREBASE_API_KEY="your-key" \
  -v $(pwd)/serviceAccountKey.json:/app/serviceAccountKey.json \
  polling-system
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DB_DSN` | `postgresql://localhost:5432/polling?sslmode=disable` | PostgreSQL connection string |
| `FIREBASE_CREDENTIALS_PATH` | `./serviceAccountKey.json` | Path to Firebase service account JSON |
| `FIREBASE_API_KEY` | *(empty)* | Firebase web API key for email/password login |
| `UPLOAD_DIR` | `./uploads` | Directory for CSV file uploads |

## API Endpoints

All routes prefixed with `/api/v1`. Swagger UI available at `/swagger/index.html`.

### Health & Docs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check |
| GET | `/swagger/*any` | Swagger UI |

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/login` | Public | Firebase login |
| POST | `/auth/token` | Bearer | Token exchange |
| GET | `/users/me` | Bearer | Get current user |
| PATCH | `/users/me` | Bearer | Update current user |

### User Management (Admin)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/users` | List users |
| GET | `/users/:id` | Get user |
| POST | `/users` | Create user |
| PATCH | `/users/:id` | Update user |
| PATCH | `/users/:id/role` | Update user role |
| DELETE | `/users/:id` | Delete user |

### Polls

| Method | Path | Role | Description |
|--------|------|------|-------------|
| POST | `/polls` | admin, poller | Create poll |
| GET | `/polls` | any | List polls (paginated) |
| GET | `/polls/:id` | any | Get poll |
| PATCH | `/polls/:id` | admin, poller | Update poll |
| DELETE | `/polls/:id` | admin | Delete poll |

### Candidates

| Method | Path | Role | Description |
|--------|------|------|-------------|
| POST | `/polls/:id/candidates` | admin, poller | Add candidate |
| GET | `/polls/:id/candidates` | any | List poll candidates |
| POST | `/polls/:id/candidates/bulk` | admin, poller | Bulk add from phone list |
| POST | `/polls/:id/candidates/upload` | admin, poller | Bulk add from CSV upload |
| GET | `/candidates?poll_id=xxx` | any | Query candidates by poll |
| GET | `/candidates/:id` | any | Get candidate |
| PATCH | `/candidates/:id` | admin, poller | Update candidate |
| DELETE | `/candidates/:id` | admin, poller | Delete candidate |
| GET | `/candidates/:id/answers` | any | List candidate answers |

**Bulk create** (`POST /polls/:id/candidates/bulk`):
```json
{ "phone_numbers": ["0412345678", "0498765432"] }
```
Response: `{ "created": 2, "duplicates": [], "invalid": [] }`

**CSV upload** (`POST /polls/:id/candidates/upload`): multipart form with `file` field. CSV must have a `phone` / `phone_number` / `phonenumber` column.

### Questions

| Method | Path | Role | Description |
|--------|------|------|-------------|
| POST | `/polls/:id/questions` | admin, poller | Create question |
| GET | `/polls/:id/questions` | any | List poll questions |
| GET | `/questions/:id` | any | Get question |
| PATCH | `/questions/:id` | admin, poller | Update question |
| DELETE | `/questions/:id` | admin, poller | Delete question |
| POST | `/questions/:id/answers` | admin, poller | Create answer |
| GET | `/questions/:id/answers` | any | List question answers |

### Answers

| Method | Path | Role | Description |
|--------|------|------|-------------|
| GET | `/answers/:id` | any | Get answer |
| DELETE | `/answers/:id` | admin | Delete answer |

### Files & Contacts

| Method | Path | Role | Description |
|--------|------|------|-------------|
| POST | `/files/upload` | admin, poller | Upload CSV of contacts |
| GET | `/files` | any | List uploaded files |
| POST | `/contacts` | admin, poller | Create contact |
| POST | `/contacts/bulk` | admin, poller | Bulk create contacts |
| GET | `/contacts` | any | List contacts (filterable) |
| PATCH | `/contacts/:id` | admin, poller | Update contact |
| DELETE | `/contacts/:id` | admin | Delete contact |

## Database Schema

7 tables with auto-migration via GORM:

- **users** — Firebase-linked application users
- **polls** — Polling campaigns
- **candidates** — People targeted by polls
- **poll_questions** — Questions within polls (supports branching via `next_question_id`)
- **answers** — Recorded answers with source tracking
- **contacts** — Contact database with phone normalization
- **files** — Uploaded CSV file metadata

Custom enums: `user_role` (admin/poller/viewer), `poll_status`, `voice_poll_method`, `answer_type`, `answer_source`.

Full schema details in [init_db_schema.md](init_db_schema.md).

## Authentication & Authorization

1. Users authenticate via `POST /auth/login` with Firebase credentials
2. All protected routes require `Authorization: Bearer <token>` header
3. Users are auto-provisioned on first request (default role: `viewer`)
4. Role-based access: `admin` > `poller` > `viewer`
5. Poll owners (pollers) can only modify their own polls and nested resources

## Key Conventions

- Paginated responses: `{ items, total, page, size, pages }`
- Error responses: `{ "error": "message" }`
- Phone numbers normalized to E.164 (Australian default: 9-digit → +61)
- CSV batch inserts in groups of 1,000 within transactions
- UUID primary keys generated via GORM `BeforeCreate` hooks

## Regenerate Swagger

```bash
swag init -g cmd/main.go -o docs/ --parseDependency --parseInternal
```

## Run Tests

```bash
go test ./tests/...
```
