
# DeskFlow — Concurrent Support Ticketing & Real-Time Routing Engine

A high-throughput, low-latency support ticketing backend engine engineered with **Go (Golang)**, **Gin**, **PostgreSQL (GORM)**, **WebSockets**, and **Docker**.

---

## 📌 What is DeskFlow?

**DeskFlow** is a backend system designed for customer support environments handling high concurrent ticket updates and real-time live agent interactions. It provides a deterministic, state-machine-driven lifecycle for tickets while offering bidirectional real-time event broadcasting to keep agents and users instantly synced.

---

## 🎯 Problems DeskFlow Solves

### 1. Concurrent State Overwrites & Race Conditions
* **The Problem:** In standard ticketing backends, when multiple support agents or automated services update a ticket status simultaneously, concurrent read-modify-write operations lead to stale data overwrites and invalid status transitions.
* **The Solution:** DeskFlow enforces strict status progression (`open` → `in_progress` → `closed`) utilizing **PostgreSQL row-level pessimistic locking (`SELECT FOR UPDATE`)** inside database transactions, preventing state conflicts and race conditions. Closed tickets cannot be reopened.

### 2. High Polling Overhead for Real-Time Updates
* **The Problem:** Traditional REST polling creates massive HTTP overhead, server load, and latency delays when multiple agents check for ticket responses and status updates.
* **The Solution:** Implements a centralized **bidirectional WebSocket hub** using Go’s native concurrency primitives (goroutines and buffered channels) to broadcast real-time events with zero polling overhead and instant client updates.

### 3. I/O Blocking During Secondary Operations
* **The Problem:** Synchronous execution of side tasks (like sending email alerts or updating audit logs) increases HTTP response latency.
* **The Solution:** Offloads notification delivery and audit events to **non-blocking worker pools**, keeping core database transactions fast, isolated, and responsive.

---

## 🏗️ System Architecture

DeskFlow follows a clean, interface-driven layered architecture:

```text
HTTP / WebSocket Request
        │
        ▼
   Router (routes/)
        │
        ▼
Middleware (middleware/)        ← JWT Authentication & Request Logging
        │
        ▼
  Handler (handlers/)           ← Request DTO parsing & validation
        │
        ▼
  Service (services/)           ← Business rules, state machine transitions
        │
        ▼
Repository (repository/)        ← Database operations & pessimistic locking
        │
        ▼
Database (PostgreSQL via GORM)

```

### Architectural Principles:

* **Separation of Concerns:** Handlers, domain services, and repository layers are strictly decoupled.
* **Interface-Driven Dependency Injection:** Mockable services and repositories injected via constructors.
* **Stateless JWT Security:** Token-based authentication with bcrypt-hashed credentials.

---

## 🗂️ Project Structure

```text
deskflow/
├── cmd/
│   └── server/
│       └── main.go              # Entry point, dependency wiring, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go            # Environment variable loading & validation
│   ├── database/
│   │   └── database.go          # DB connection & GORM auto-migrations
│   ├── handlers/
│   │   ├── auth_handler.go      # User registration & login endpoints
│   │   ├── ticket_handler.go    # Ticket CRUD & status mutation handlers
│   │   └── health_handler.go    # System health check endpoint
│   ├── middleware/
│   │   └── auth.go              # JWT authentication & route gating
│   ├── models/
│   │   ├── user.go              # User models & auth DTOs
│   │   └── ticket.go            # Ticket models, state constants & DTOs
│   ├── repository/
│   │   ├── user_repository.go   # User data access layer
│   │   └── ticket_repository.go # Ticket queries & transactional updates
│   ├── routes/
│   │   └── routes.go            # Central API route definitions
│   └── services/
│       ├── auth_service.go      # Authentication & token generation logic
│       └── ticket_service.go    # Ticket lifecycle & state enforcement
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # App + PostgreSQL container orchestration
├── .env.example                 # Environment configuration template
├── go.mod
└── README.md

```

---

## 🚀 API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/health` | Health check endpoint |
| `POST` | `/auth/register` | Register a new user |
| `POST` | `/auth/login` | Authenticate user and receive JWT token |

### Protected Endpoints (`Authorization: Bearer <token>`)

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/tickets` | Create a new support ticket |
| `GET` | `/tickets` | List tickets belonging to authenticated user |
| `GET` | `/tickets/:id` | Fetch details of a specific ticket (owner gated) |
| `PATCH` | `/tickets/:id/status` | Update ticket status (`open` → `in_progress` → `closed`) |

---

## ⚙️ Environment Configuration

Create a `.env` file in the root directory:

```bash
cp .env.example .env

```

| Variable | Description | Default |
| --- | --- | --- |
| `PORT` | Application server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | Database name | `ticketdb` |
| `JWT_SECRET` | Secret key for JWT signing | *required* |
| `JWT_EXPIRY_HOURS` | Token validity window | `24` |

---

## 🐳 Getting Started

### Using Docker Compose (Recommended)

Run the full stack (Go application + PostgreSQL) in one command:

```bash
# Build and run containers
docker-compose up --build -d

# View live logs
docker-compose logs -f app

# Stop services
docker-compose down

```

### Local Development

```bash
# Install Go dependencies
go mod download

# Run database migrations and start server
go run cmd/server/main.go

```

Verify the service is running:

```bash
curl http://localhost:8080/health
# Output: {"status":"ok"}

```

```

```
