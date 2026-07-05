# Chat App Backend

Online chat application backend built with Go, GoFr framework, and PostgreSQL.

## Tech Stack

- **Go** 1.21+
- **GoFr** v1.x — HTTP framework, dependency injection, config management
- **PostgreSQL** 15+ — Primary database
- **gorilla/websocket** — Real-time WebSocket communication
- **golang-jwt/jwt** v5 — JWT authentication
- **bcrypt** — Password hashing

## Quick Start

### 1. Start Dependencies

```bash
docker-compose up -d postgres redis
```

### 2. Run Database Migrations

```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgres://chat:chat123@localhost:5432/chatapp?sslmode=disable" up
```

### 3. Configure Environment

Copy the example config and adjust as needed:

```bash
cp configs/.env.example configs/.env
```

### 4. Run the Application

```bash
go run cmd/server/main.go
```

The server starts on port 8000 by default.

## API Overview

### Authentication
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/login` | User login |

### User (requires Bearer token)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users/profile` | Get current user profile |
| PUT | `/api/v1/users/profile` | Update profile |
| GET | `/api/v1/users/search?keyword=xxx` | Search users |

### WebSocket
| Path | Description |
|------|-------------|
| `ws://host:8000/ws?token=<jwt>` | WebSocket connection |

## Configuration

All configuration is via environment variables (or `configs/.env`):

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | 8000 | Server port |
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | chat | Database user |
| `DB_PASSWORD` | chat123 | Database password |
| `DB_NAME` | chatapp | Database name |
| `JWT_SECRET` | — | JWT signing secret (change in production!) |
| `JWT_EXPIRE_HOURS` | 72 | Token expiry in hours |

## Development

```bash
# Build
go build -o bin/chat-app ./cmd/server

# Test
go test -v -race -cover ./...

# Lint
golangci-lint run
```

## Project Structure

```
cmd/server/         — Application entry point
internal/
  handler/          — HTTP request handlers
  service/          — Business logic layer
  repository/       — Data access layer
  model/            — Data models
  middleware/       — Auth, CORS middleware
  ws/               — WebSocket hub and client
  router/           — Route definitions
  pkg/              — Shared utilities (JWT, hash, response, error codes)
migrations/         — Database migration SQL
configs/            — Environment configuration
```
