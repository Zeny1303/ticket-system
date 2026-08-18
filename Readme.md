# DeskFlow — Concurrent Support Ticketing & Real-Time Routing Engine

A high-throughput, low-latency support ticketing backend built with Go (Gin), PostgreSQL (GORM), WebSockets, and Docker. DeskFlow is designed for customer-support systems that require deterministic ticket state transitions, safe concurrent updates, and real-time agent notifications.

Badges: <!-- add CI/coverage/license badges here -->

---

## Quick overview

- Purpose: Provide a robust backend for support teams that need safe concurrent ticket updates and real-time agent coordination.
- Built with: Go, Gin, GORM (Postgres), WebSockets, Docker.
- Key concepts: pessimistic row-level locking for state transitions, a central WebSocket hub for broadcasts, and background worker pools for non-blocking side-effects.

---

## Features

- Deterministic ticket state machine: `open` → `in_progress` → `closed` with enforced transitions
- Safe concurrent updates using PostgreSQL `SELECT FOR UPDATE` within transactions
- Real-time two-way updates via a centralized WebSocket hub
- Non-blocking worker pools for notifications and audit logging
- JWT auth with bcrypt-hashed credentials
- Docker & docker-compose for easy local setup

---

## Quickstart

Recommended: Docker Compose (app + Postgres)

```bash
# Copy env template and start stack
cp .env.example .env
docker-compose up --build -d

docker-compose logs -f app
# Stop
# docker-compose down
```

Local development

```bash
# Install deps
go mod download

# Run migrations and start server
go run cmd/server/main.go

# Health-check
curl http://localhost:8080/health
# => {"status":"ok"}
```

---

## Environment (create a `.env` from `.env.example`)

| Variable | Description | Default / Required |
| --- | --- | --- |
| PORT | Application listen port | 8080 |
| DB_HOST | Postgres host | localhost |
| DB_PORT | Postgres port | 5432 |
| DB_USER | Postgres user | postgres |
| DB_PASSWORD | Postgres password | postgres |
| DB_NAME | Database name | ticketdb |
| JWT_SECRET | Secret used to sign JWT tokens | required |
| JWT_EXPIRY_HOURS | JWT token lifetime in hours | 24 |

---

## API

All protected endpoints require: `Authorization: Bearer <token>`

Public endpoints

- GET /health — health check
- POST /auth/register — register user
- POST /auth/login — returns JWT token

Protected endpoints (examples)

- POST /tickets — create ticket
  curl -H "Authorization: Bearer <token>" -H "Content-Type: application/json" \
    -d '{"title":"Issue","body":"Details"}' http://localhost:8080/tickets

- GET /tickets — list tickets for authenticated user

- PATCH /tickets/:id/status — update status (must follow progression)

Notes

- Status transitions are validated in the service layer and performed inside DB transactions with row-level locking to avoid race conditions.
- Real-time updates: agents/clients can connect to the WebSocket hub to receive ticket events and broadcasts.

---

## System architecture

DeskFlow uses a layered, interface-driven architecture to separate concerns and make the codebase testable.

High-level flow:

HTTP / WebSocket request → Router → Middleware (JWT, logging) → Handlers (DTO parsing) → Services (business logic, state machine) → Repositories (DB operations, transactions) → PostgreSQL (GORM)

Principles

- Separation of concerns
- Interface-driven dependency injection for easy testing
- Stateless JWT auth for horizontal scalability

---

## Project layout

```
cmd/server/main.go          # entrypoint, dependency wiring
internal/config             # env loading
internal/database           # DB connection & migrations
internal/handlers           # HTTP + WS handlers
internal/services           # business logic & state machine
internal/repository         # DB layer (pessimistic locks)
internal/middleware         # auth + logging
internal/models             # User and Ticket models
Dockerfile
docker-compose.yml
.env.example
```

---

## WebSocket hub

The app exposes a WebSocket hub (see handlers/ws) that maintains client connections and broadcasts ticket events. The hub is built around buffered channels and goroutines to ensure low-latency broadcasts and backpressure handling.

---

## Testing & development tips

- Use the provided docker-compose.yml to run Postgres locally for development.
- Services and repositories are interface-driven; write unit tests by providing mocks.
- Keep long-running side-effects out of DB transactions — use worker pools for notifications and audit logging.

---

## Contributing

Contributions welcome. Suggested workflow:

1. Fork the repo
2. Create a feature branch
3. Add tests for new behavior
4. Open a PR describing the change

Please follow Go idioms and add unit tests for service and repository behavior.

---

## License

Specify your chosen license here (e.g. MIT). If you have a license file, link to it.

---

If you'd like, I can:
- Add example curl requests for the auth and ticket flows
- Add a small WebSocket client example (JS)
- Add CI/coverage and repository badges

