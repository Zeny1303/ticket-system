# Ticket System — Backend API

A production-quality REST API for a ticket management system built with Go, Gin, GORM, and PostgreSQL.

---

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [Folder Structure](#folder-structure)
- [API Endpoints](#api-endpoints)
- [Database Schema](#database-schema)
- [Environment Setup](#environment-setup)
- [Local Setup (without Docker)](#local-setup-without-docker)
- [Docker Instructions](#docker-instructions)
- [Deployment on Railway](#deployment-on-railway)
- [API Testing](#api-testing)
- [Assumptions](#assumptions)
- [Future Improvements](#future-improvements)
- [Development Timeline](#development-timeline)

---

## Project Overview

A backend service where users can:
- Register and login with JWT authentication
- Create support tickets
- View only their own tickets
- Update ticket status following a strict state machine: `open → in_progress → closed`
- A closed ticket cannot be reopened

---

## Architecture

```
HTTP Request
     │
     ▼
  Router (routes/)
     │
     ▼
Middleware (middleware/)   ← JWT validation on protected routes
     │
     ▼
  Handler (handlers/)      ← Parse request, validate input, call service
     │
     ▼
  Service (services/)      ← Business logic, business rules
     │
     ▼
Repository (repository/)   ← Database queries only
     │
     ▼
  Database (PostgreSQL via GORM)
```

**Design Principles Applied:**
- **Separation of Concerns** — each layer has one job
- **Dependency Injection** — dependencies are injected via constructors, not globals
- **Repository Pattern** — data access is isolated from business logic
- **Interface-driven design** — services and repositories are defined as interfaces

---

## Folder Structure

```
ticket-system/
├── cmd/
│   └── server/
│       └── main.go          # Entry point, dependency wiring, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go        # Environment variable loading and validation
│   ├── database/
│   │   └── database.go      # PostgreSQL connection and auto-migration
│   ├── handlers/
│   │   ├── auth_handler.go  # POST /auth/register, POST /auth/login
│   │   ├── ticket_handler.go # All /tickets endpoints
│   │   └── health_handler.go # GET /health
│   ├── middleware/
│   │   └── auth.go          # JWT authentication middleware
│   ├── models/
│   │   ├── user.go          # User struct + request/response DTOs
│   │   └── ticket.go        # Ticket struct + status constants + DTOs
│   ├── repository/
│   │   ├── user_repository.go   # User DB queries
│   │   └── ticket_repository.go # Ticket DB queries
│   ├── routes/
│   │   └── routes.go        # URL routing table
│   └── services/
│       ├── auth_service.go  # Registration and login business logic
│       └── ticket_service.go # Ticket CRUD and status transition logic
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yml       # App + PostgreSQL orchestration
├── .env.example             # Environment variable template
├── .gitignore
├── go.mod
└── README.md
```

---

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check — returns `{"status": "ok"}` |
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Login and receive JWT token |

### Protected Endpoints (require `Authorization: Bearer <token>`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/tickets` | Create a new ticket |
| GET | `/tickets` | List all tickets for authenticated user |
| GET | `/tickets/:id` | Get a specific ticket (owner only) |
| PATCH | `/tickets/:id/status` | Update ticket status |

### Request / Response Examples

**POST /auth/register**
```json
// Request
{
  "email": "user@example.com",
  "password": "secret123"
}

// Response 201
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  }
}
```

**POST /tickets**
```json
// Request (with Authorization: Bearer <token>)
{
  "title": "Login page is broken",
  "description": "Users cannot log in on mobile devices"
}

// Response 201
{
  "success": true,
  "message": "Ticket created successfully",
  "data": {
    "id": 1,
    "title": "Login page is broken",
    "description": "Users cannot log in on mobile devices",
    "status": "open",
    "user_id": 1,
    "created_at": "2024-01-15T10:05:00Z",
    "updated_at": "2024-01-15T10:05:00Z"
  }
}
```

**PATCH /tickets/:id/status**
```json
// Request
{
  "status": "in_progress"
}

// Response 200
{
  "success": true,
  "message": "Ticket status updated successfully",
  "data": { ... }
}

// Error — invalid transition
{
  "success": false,
  "message": "invalid status transition. Allowed flow: open -> in_progress -> closed. Closed tickets cannot be reopened",
  "data": null
}
```

---

## Database Schema

```sql
-- users table
CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    email      VARCHAR UNIQUE NOT NULL,
    password   VARCHAR NOT NULL,          -- bcrypt hash, never plain text
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

-- tickets table
CREATE TABLE tickets (
    id          SERIAL PRIMARY KEY,
    title       VARCHAR NOT NULL,
    description TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'open',
    user_id     INTEGER NOT NULL REFERENCES users(id),
    created_at  TIMESTAMP WITH TIME ZONE,
    updated_at  TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tickets_user_id ON tickets(user_id);
```

> Tables are created automatically by GORM's AutoMigrate on startup.

---

## Environment Setup

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | — |
| `DB_NAME` | Database name | `ticketdb` |
| `JWT_SECRET` | JWT signing secret (**required**) | — |
| `JWT_EXPIRY_HOURS` | Token validity in hours | `24` |

Generate a secure JWT secret:
```bash
openssl rand -hex 32
```

---

## Local Setup (without Docker)

**Prerequisites:** Go 1.22+, PostgreSQL 14+

```bash
# 1. Clone the repository
git clone https://github.com/Zeny1303/ticket-system.git
cd ticket-system

# 2. Install dependencies
go mod download

# 3. Create the database
psql -U postgres -c "CREATE DATABASE ticketdb;"

# 4. Set up environment variables
cp .env.example .env
# Edit .env with your PostgreSQL credentials and a strong JWT_SECRET

# 5. Run the application
go run ./cmd/server

# Server starts on http://localhost:8080
```

---

## Docker Instructions

### Option A: Docker Compose (Recommended — runs app + PostgreSQL together)

```bash
# Build and start everything
docker-compose up --build

# Run in background
docker-compose up --build -d

# View logs
docker-compose logs -f app

# Stop everything
docker-compose down

# Stop and delete all data (including PostgreSQL volume)
docker-compose down -v
```

### Option B: Docker only (requires external PostgreSQL)

```bash
# Build the image
docker build -t ticket-system .

# Run the container
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=yourpassword \
  -e DB_NAME=ticketdb \
  -e JWT_SECRET=your-secret-key \
  ticket-system
```

### Verify the deployment

```bash
curl http://localhost:8080/health
# Expected: {"status":"ok"}
```

---

## Deployment on Railway

Railway is a free-tier platform that supports Go and Docker deployments.

### Steps

1. **Create account** at [railway.app](https://railway.app)

2. **Create a new project** → "Deploy from GitHub repo"

3. **Add PostgreSQL** → Click "+ New" → "Database" → "PostgreSQL"
   - Railway automatically sets `DATABASE_URL` — we don't use this format, so set individual variables below

4. **Set environment variables** in Railway dashboard → Variables:
   ```
   PORT=8080
   DB_HOST=<from Railway PostgreSQL service>
   DB_PORT=5432
   DB_USER=<from Railway PostgreSQL service>
   DB_PASSWORD=<from Railway PostgreSQL service>
   DB_NAME=railway
   JWT_SECRET=<generate with: openssl rand -hex 32>
   JWT_EXPIRY_HOURS=24
   ```
   > Get DB credentials from: Railway project → PostgreSQL service → Variables tab

5. **Deploy** — Railway detects the Dockerfile automatically and builds it.

6. **Get your public URL** from Railway dashboard → Settings → Domains → "Generate Domain"

7. **Test deployment:**
   ```bash
   curl https://your-app.railway.app/health
   ```

---

## API Testing

### Using cURL

```bash
# Health check
curl http://localhost:8080/health

# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Login (save the token from response)
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# Set token variable (replace with actual token)
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Create ticket
curl -X POST http://localhost:8080/tickets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Bug in login page","description":"Users cannot log in on Safari"}'

# List tickets
curl http://localhost:8080/tickets \
  -H "Authorization: Bearer $TOKEN"

# Get ticket by ID
curl http://localhost:8080/tickets/1 \
  -H "Authorization: Bearer $TOKEN"

# Update status to in_progress
curl -X PATCH http://localhost:8080/tickets/1/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"status":"in_progress"}'

# Update status to closed
curl -X PATCH http://localhost:8080/tickets/1/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"status":"closed"}'

# Try to reopen closed ticket — should fail with 422
curl -X PATCH http://localhost:8080/tickets/1/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"status":"open"}'
```

### Using Postman

1. Import a new collection named "Ticket System"
2. Set a collection variable: `base_url = http://localhost:8080`
3. Set a collection variable: `token = ` (fill after login)
4. For protected requests: Authorization tab → Bearer Token → `{{token}}`

**Test sequence:**
1. `GET {{base_url}}/health`
2. `POST {{base_url}}/auth/register` — save token to collection variable
3. `POST {{base_url}}/tickets` — create a ticket, save ticket ID
4. `GET {{base_url}}/tickets` — list tickets
5. `GET {{base_url}}/tickets/1`
6. `PATCH {{base_url}}/tickets/1/status` → `{"status":"in_progress"}`
7. `PATCH {{base_url}}/tickets/1/status` → `{"status":"closed"}`
8. `PATCH {{base_url}}/tickets/1/status` → `{"status":"open"}` — expect 422 error

---

## Assumptions

1. **Auto-login on register** — A JWT token is returned immediately upon registration. The client does not need to make a separate login call.
2. **Status initial value** — All new tickets start with status `open`. The client cannot choose the initial status.
3. **Ownership on GET** — Requesting a ticket that doesn't belong to the user returns 404, not 403. This prevents leaking information about the existence of other users' tickets.
4. **Empty ticket list** — Returns `{"data": []}` not `{"data": null}` for users with no tickets.
5. **No pagination** — The assignment does not require it; all user tickets are returned in a single response.
6. **No refresh tokens** — Only access tokens are used. Token expiry is configurable via `JWT_EXPIRY_HOURS`.

---

## Future Improvements

1. **Pagination** — Add `page` and `limit` query parameters to `GET /tickets`
2. **Refresh tokens** — Implement a token refresh endpoint to avoid forcing re-login
3. **Structured logging** — Replace `log.Printf` with `zerolog` or `zap` for JSON-structured logs
4. **Database migrations** — Replace GORM AutoMigrate with `golang-migrate` for production-safe migrations
5. **Unit tests** — Test service layer with mock repositories using `testify`
6. **Rate limiting** — Add `golang.org/x/time/rate` middleware to prevent brute force on auth endpoints
7. **Request ID** — Add a unique request ID to every request for distributed tracing
8. **Ticket filtering** — Filter tickets by status via query parameter: `GET /tickets?status=open`

---

## Development Timeline

### Day 1
- 09:00 — Project setup, go mod init, folder structure
- 09:45 — Dependencies installation (Gin, GORM, JWT, bcrypt, validator)
- 10:30 — Config layer (environment variable loading)
- 11:00 — Models (User, Ticket, DTOs, status constants)
- 11:30 — Database connection and auto-migration
- 12:00 — Utils (JWT generation/validation, password hashing, response helpers)
- 13:30 — Repository layer (UserRepository, TicketRepository)
- 15:00 — Service layer (AuthService, TicketService with business rules)
- 16:30 — Auth handlers (Register, Login)
- 17:30 — JWT middleware

### Day 2
- 09:00 — Ticket handlers (Create, List, GetByID, UpdateStatus)
- 10:00 — Route registration
- 10:30 — main.go (wiring, graceful shutdown)
- 11:30 — Dockerfile (multi-stage build)
- 12:00 — docker-compose.yml
- 13:00 — Local testing with Docker Compose
- 14:00 — Deployment on Railway
- 15:00 — README documentation
- 16:00 — Final testing of all endpoints
- 17:00 — Submission