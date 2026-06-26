# Ticket System (Golang Backend Internship Assignment)

A small backend service for a ticket system. Users can register, log in,
create tickets, view only their own tickets, and move a ticket through a
fixed status lifecycle: `open -> in_progress -> closed`.

## Tech Stack

- **Language:** Go 1.22
- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **ORM / DB:** [GORM](https://gorm.io) + SQLite (pure-Go driver, no CGO required)
- **Auth:** JWT (`github.com/golang-jwt/jwt/v5`), passwords hashed with bcrypt

## Project Structure

```
ticket-system/
├── cmd/
│   └── server/
│       └── main.go            # Application entrypoint
├── internal/
│   ├── config/                # Env var loading / app config
│   │   └── config.go
│   ├── database/               # DB connection + migrations
│   │   └── database.go
│   ├── models/                 # GORM models + business rules
│   │   ├── user.go
│   │   └── ticket.go
│   ├── middleware/              # JWT auth middleware
│   │   └── auth.go
│   ├── utils/                   # JWT, bcrypt, response helpers
│   │   ├── jwt.go
│   │   ├── password.go
│   │   └── response.go
│   ├── handlers/                # Request handlers / controllers
│   │   ├── auth_handler.go
│   │   └── ticket_handler.go
│   └── routes/
│       └── routes.go            # Route registration
├── go.mod
├── .env.example
├── Dockerfile
├── .dockerignore
└── README.md
```

## Environment Variables

Copy `.env.example` to `.env` and adjust as needed:

| Variable           | Description                              | Default                          |
|---------------------|-------------------------------------------|-----------------------------------|
| `PORT`              | HTTP port the server listens on           | `8080`                            |
| `DATABASE_PATH`     | SQLite file path                          | `ticket_system.db`                |
| `JWT_SECRET`        | Secret used to sign/verify JWTs           | `change-this-secret-in-production`|
| `JWT_EXPIRY_HOURS`  | JWT validity window, in hours             | `24`                               |
| `GIN_MODE`          | `release` or `debug`                      | `release`                          |

## Running Locally (without Docker)

```bash
cp .env.example .env
go mod tidy
go run ./cmd/server
curl http://localhost:8080/health
```

## Running with Docker (required)

```bash
docker build -t ticket-system .
docker run -p 8080:8080 ticket-system
curl http://localhost:8080/health
```

Expected health response:

```json
{ "status": "ok" }
```

To persist the SQLite file across container restarts, mount a volume:

```bash
docker run -p 8080:8080 -v $(pwd)/data:/app/data \
  -e DATABASE_PATH=/app/data/ticket_system.db \
  ticket-system
```

## API Reference

All responses are JSON. Protected endpoints require:

```
Authorization: Bearer <token>
```

### `GET /health`

Public health check.

**Response `200`**
```json
{ "status": "ok" }
```

---

### `POST /auth/register`

Registers a new user. Passwords are hashed with bcrypt before storage.

**Request**
```json
{ "email": "user@example.com", "password": "secret123" }
```

**Response `201`**
```json
{ "id": 1, "email": "user@example.com", "created_at": "2026-06-25T10:00:00Z" }
```

**Response `400`** — invalid payload (e.g. bad email format, password too short)
```json
{ "error": "invalid request payload: ..." }
```

**Response `409`** — duplicate email
```json
{ "error": "email already registered" }
```

---

### `POST /auth/login`

**Request**
```json
{ "email": "user@example.com", "password": "secret123" }
```

**Response `200`**
```json
{ "token": "eyJhbGciOiJIUzI1NiIs...", "token_type": "Bearer", "expires_in_hours": 24 }
```

**Response `401`** — wrong email or password
```json
{ "error": "invalid email or password" }
```

---

### `POST /tickets` (protected)

Creates a ticket owned by the authenticated user. New tickets always start
in `open` status.

**Request**
```json
{ "title": "Server down", "description": "Production API returning 500s" }
```

**Response `201`**
```json
{
  "id": 1,
  "title": "Server down",
  "description": "Production API returning 500s",
  "status": "open",
  "user_id": 1,
  "created_at": "2026-06-25T10:00:00Z",
  "updated_at": "2026-06-25T10:00:00Z"
}
```

**Response `401`** — missing/invalid token
```json
{ "error": "missing Authorization header" }
```

---

### `GET /tickets` (protected)

Lists only the tickets owned by the authenticated user.

**Response `200`**
```json
{
  "tickets": [
    {
      "id": 1,
      "title": "Server down",
      "description": "Production API returning 500s",
      "status": "open",
      "user_id": 1,
      "created_at": "2026-06-25T10:00:00Z",
      "updated_at": "2026-06-25T10:00:00Z"
    }
  ]
}
```

---

### `GET /tickets/:id` (protected)

Returns a single ticket — only if it belongs to the authenticated user.

**Response `200`**
```json
{
  "id": 1,
  "title": "Server down",
  "description": "Production API returning 500s",
  "status": "open",
  "user_id": 1,
  "created_at": "2026-06-25T10:00:00Z",
  "updated_at": "2026-06-25T10:00:00Z"
}
```

**Response `404`** — ticket does not exist
```json
{ "error": "ticket not found" }
```

**Response `403`** — ticket exists but belongs to another user
```json
{ "error": "you do not have access to this ticket" }
```

---

### `PATCH /tickets/:id/status` (protected)

Updates a ticket's status, enforcing the forward-only flow
`open -> in_progress -> closed`. A closed ticket can never be reopened or
moved to any other status.

**Request**
```json
{ "status": "in_progress" }
```

**Response `200`**
```json
{ "id": 1, "status": "in_progress", "updated_at": "2026-06-25T10:05:00Z" }
```

**Response `400`** — invalid status value
```json
{ "error": "invalid status: must be one of open, in_progress, closed" }
```

**Response `400`** — invalid transition (e.g. `open` -> `closed` directly, or any change on a closed ticket)
```json
{ "error": "cannot transition ticket from 'open' to 'closed'" }
```
```json
{ "error": "closed ticket cannot be reopened or modified" }
```

**Response `403`** — ticket belongs to another user
```json
{ "error": "you do not have access to this ticket" }
```

**Response `404`** — ticket not found
```json
{ "error": "ticket not found" }
```

## HTTP Status Codes Used

| Code | Meaning                                                |
|------|----------------------------------------------------------|
| 200  | Successful read/update                                    |
| 201  | Resource created (register, ticket creation)              |
| 400  | Bad request / validation error / invalid status transition|
| 401  | Missing, invalid, or expired JWT; bad login credentials   |
| 403  | Authenticated but not the owner of the resource            |
| 404  | Resource not found                                         |
| 409  | Conflict (duplicate email on registration)                  |
| 500  | Unexpected server error                                     |

## Design Notes / Assumptions

- Ticket statuses are strictly forward-only: `open -> in_progress -> closed`.
  Setting a ticket to its current status, skipping a step (`open` directly
  to `closed`), or modifying a `closed` ticket are all rejected with `400`.
- Ownership is enforced at the handler level: a ticket lookup that finds a
  ticket owned by someone else returns `403`, while a non-existent ticket
  returns `404`.
- The SQLite driver used (`github.com/glebarez/sqlite`) is pure Go, so the
  Docker image does not need a C compiler/CGO — keeping the image small and
  the build simple and portable across any free-tier Go/Docker host.
- No admin role, ticket assignment, or comments module is implemented, per
  the assignment scope.
- Emails are normalized to lowercase before storage/lookup to avoid
  case-sensitivity duplicate-account issues.

## Deployment

This service is a single statically-linked binary in a small Alpine image
listening on port `8080`, so it can be deployed to any free-tier platform
that supports Docker, such as Render, Fly.io, Railway, or Koyeb. After
deploying:

1. Confirm `GET /health` is publicly reachable and returns `{"status": "ok"}`.
2. Set `JWT_SECRET` to a strong, unique value via the platform's environment
   variable settings (do not rely on the default).
3. If the platform's filesystem is ephemeral, note that the SQLite database
   resets on redeploy/restart unless a persistent volume is attached.

**Deployed URL:** _add after deployment_
**Public health check URL:** _add after deployment_ (e.g. `https://<your-app>/health`)
