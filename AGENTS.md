# AGENTS.md

## Project Overview

Chat application backend built with GoFr framework and PostgreSQL. Module name: `chatgoo`.

## Architecture

Three-layer architecture: **Handler → Service → Repository**

- **Handler** (`internal/handler/`): HTTP request/response handling. Handlers are closures that capture service dependencies. Return `(any, error)` — GoFr handles JSON serialization and status codes.
- **Service** (`internal/service/`): Business logic. Defines interfaces for testability. Validates input, orchestrates repository calls, manages side effects.
- **Repository** (`internal/repository/`): Data access. Uses GoFr's `container.DB` interface for PostgreSQL. All SQL uses parameterized queries (`$1, $2`).

## Key Patterns

### Dependency Injection
GoFr v1.x does NOT have `AddDependency`/`Get`. Use closures instead:
```go
func MyHandler(svc service.MyService) func(c *gofr.Context) (any, error) {
    return func(c *gofr.Context) (any, error) {
        // use svc here
    }
}
```

### Database Access
Repositories receive `container.DB` at construction:
```go
repo := repository.NewUserRepository(app.GetSQL())
```
The `container.DB` interface provides `QueryContext`, `QueryRowContext`, `ExecContext`, `Begin`, etc.

### Authentication
Global JWT auth middleware registered via `app.UseMiddleware()`. Public paths are skipped. User ID/username stored in request context via `context.WithValue`.

### WebSocket
Uses GoFr's `app.WebSocket()` with gorilla/websocket under the hood. Auth via `?token=` query parameter. Hub pattern for broadcasting messages to sessions.

### Config
GoFr auto-loads `.env` files from `./configs/`. Access via `app.Config.Get("KEY")`.

## Directory Structure

```
cmd/server/main.go          — Entry point
internal/
  handler/                   — HTTP handlers (closures)
  service/                   — Business logic (interfaces + implementations)
  repository/                — Database access (container.DB)
  model/                     — Data structs with json/db tags
  middleware/                 — HTTP middleware (auth, CORS)
  ws/                        — WebSocket hub, client, handler
  router/                    — Route registration
  pkg/errcode/               — Error code constants
  pkg/response/              — Unified API response helpers
  pkg/jwt/                   — JWT generate/parse
  pkg/hash/                  — bcrypt hash/verify
migrations/                  — SQL migration scripts
configs/                     — .env configuration
```

## Conventions

- Package names: lowercase, singular (`model`, not `models`)
- File names: lowercase with underscores (`user_repo.go`)
- Error variables: `Err` prefix (`ErrNotFound`)
- Constants: PascalCase (`MessageTypeText`)
- All SQL: parameterized (`$1`, `$2`), never string concatenation
- Time fields: `TIMESTAMPTZ` everywhere
- Primary keys: `BIGSERIAL`

## Adding a New Entity

1. Create model in `internal/model/`
2. Create repository interface + implementation in `internal/repository/`
3. Create service interface + implementation in `internal/service/`
4. Create handler closures in `internal/handler/`
5. Register routes in `internal/router/router.go`
6. Wire dependencies in `cmd/server/main.go`

## Running

```bash
# Start dependencies
docker-compose up -d postgres redis

# Run application
go run cmd/server/main.go

# Run tests
go test -v -race -cover ./...
```
