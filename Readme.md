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
This project was developed for the EVA BHARAT Backend Developer Assignment.
