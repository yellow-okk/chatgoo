# Chat App Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the GoFr+PostgreSQL chat application backend with database, WebSocket, and user management modules, plus project documentation.

**Architecture:** Handler → Service → Repository three-layer architecture using GoFr framework. PostgreSQL for persistence, gorilla/websocket for real-time communication, JWT for authentication.

**Tech Stack:** Go 1.21+, GoFr v1.x, PostgreSQL 15+, pgx v5, gorilla/websocket, golang-jwt/jwt v5, bcrypt, testify

## Global Constraints

- Module name: `chatgoo` (from existing go.mod)
- All SQL queries use parameterized `$1, $2` syntax — no string concatenation
- Time fields use `TIMESTAMPTZ`; primary keys use `BIGSERIAL`
- All exported functions have GoDoc comments
- Error variables prefixed with `Err`; constants use PascalCase
- WebSocket heartbeat: client ping every 30s, server pong timeout 60s
- JWT tokens expire after 72 hours by default
- Password hashing uses bcrypt with cost ≥ 10

---

### Task 1: Project Scaffolding & Dependencies

**Covers:** §2, §3, §4, §6

**Files:**
- Modify: `go.mod`
- Create: `configs/.env.example`
- Create: `.gitignore`
- Create: `migrations/001_init_schema.up.sql`
- Create: `migrations/001_init_schema.down.sql`
- Create: `docker-compose.yml`
- Create: `Dockerfile`
- Create: `Makefile`

**Interfaces:**
- Produces: working `go.mod` with all dependencies; project directory structure; database migration scripts

- [ ] **Step 1: Update go.mod and fetch dependencies**

```bash
cd E:\Go\chatgoo
go get gofr.dev/pkg/gofr
go get github.com/gorilla/websocket
go get github.com/golang-jwt/jwt/v5
go get github.com/stretchr/testify
go get github.com/lib/pq
```

- [ ] **Step 2: Create .gitignore**

```gitignore
# Binaries
bin/
*.exe

# IDE
.idea/
.vscode/
*.swp

# Environment
.env
configs/.env

# Uploads
uploads/

# OS
.DS_Store
Thumbs.db
```

- [ ] **Step 3: Create configs/.env.example**

```env
APP_ENV=development
APP_NAME=chatgoo
HTTP_PORT=8000

DB_HOST=localhost
DB_PORT=5432
DB_USER=chat
DB_PASSWORD=chat123
DB_NAME=chatapp
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300s

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=your-super-secret-key-change-in-prod
JWT_EXPIRE_HOURS=72

UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=50

WS_READ_BUFFER_SIZE=10240
WS_WRITE_BUFFER_SIZE=10240
WS_PING_INTERVAL=30s
```

- [ ] **Step 4: Create migrations/001_init_schema.up.sql**

File order: files BEFORE messages (messages.file_id references files.file_id).

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    user_id          BIGSERIAL PRIMARY KEY,
    username         VARCHAR(50)  NOT NULL UNIQUE,
    password_hash    VARCHAR(255) NOT NULL,
    nickname         VARCHAR(50)  NOT NULL,
    avatar_url       VARCHAR(500),
    gender           SMALLINT     DEFAULT 0,
    signature        VARCHAR(200),
    region           VARCHAR(100),
    birthday         DATE,
    status           SMALLINT     DEFAULT 0,
    last_login_at    TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_users_gender CHECK (gender IN (0, 1, 2)),
    CONSTRAINT chk_users_status CHECK (status IN (0, 1, 2))
);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);

CREATE TABLE friend_groups (
    group_id         BIGSERIAL PRIMARY KEY,
    user_id          BIGINT       NOT NULL,
    group_name       VARCHAR(50)  NOT NULL,
    created_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_fg_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT uq_fg_user_group UNIQUE (user_id, group_name)
);
CREATE INDEX idx_fg_user ON friend_groups(user_id);

CREATE TABLE friend_relations (
    relation_id      BIGSERIAL PRIMARY KEY,
    user_id          BIGINT       NOT NULL,
    friend_id        BIGINT       NOT NULL,
    group_id         BIGINT,
    remark           VARCHAR(50),
    status           SMALLINT     DEFAULT 0,
    applied_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    approved_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_fr_user   FOREIGN KEY (user_id)   REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_fr_friend FOREIGN KEY (friend_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_fr_group  FOREIGN KEY (group_id)  REFERENCES friend_groups(group_id) ON DELETE SET NULL,
    CONSTRAINT chk_fr_status CHECK (status IN (0, 1, 2, 3)),
    CONSTRAINT chk_fr_diff CHECK (user_id <> friend_id)
);
CREATE INDEX idx_fr_user_friend ON friend_relations(user_id, friend_id);
CREATE INDEX idx_fr_friend_status ON friend_relations(friend_id, status);

CREATE TABLE group_info (
    group_id         BIGSERIAL PRIMARY KEY,
    group_name       VARCHAR(100) NOT NULL,
    owner_id         BIGINT       NOT NULL,
    avatar_url       VARCHAR(500),
    announcement     TEXT,
    max_members      INT          DEFAULT 200,
    created_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_gi_owner FOREIGN KEY (owner_id) REFERENCES users(user_id)
);
CREATE INDEX idx_gi_owner ON group_info(owner_id);

CREATE TABLE group_members (
    id               BIGSERIAL PRIMARY KEY,
    group_id         BIGINT       NOT NULL,
    user_id          BIGINT       NOT NULL,
    role             SMALLINT     DEFAULT 0,
    join_at          TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_gm_group FOREIGN KEY (group_id) REFERENCES group_info(group_id) ON DELETE CASCADE,
    CONSTRAINT fk_gm_user  FOREIGN KEY (user_id)  REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT chk_gm_role CHECK (role IN (0, 1, 2)),
    CONSTRAINT uq_gm_group_user UNIQUE (group_id, user_id)
);
CREATE INDEX idx_gm_group ON group_members(group_id);
CREATE INDEX idx_gm_user  ON group_members(user_id);

CREATE TABLE chat_sessions (
    session_id       BIGSERIAL PRIMARY KEY,
    session_type     SMALLINT     NOT NULL,
    target_user_id   BIGINT,
    group_id         BIGINT,
    last_message_id  BIGINT,
    last_message_at  TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_cs_type CHECK (session_type IN (1, 2)),
    CONSTRAINT chk_cs_target CHECK (
        (session_type = 1 AND target_user_id IS NOT NULL AND group_id IS NULL)
        OR
        (session_type = 2 AND group_id IS NOT NULL AND target_user_id IS NULL)
    ),
    CONSTRAINT fk_cs_target FOREIGN KEY (target_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_cs_group  FOREIGN KEY (group_id)      REFERENCES group_info(group_id) ON DELETE CASCADE
);
CREATE INDEX idx_cs_type_target ON chat_sessions(session_type, target_user_id);
CREATE INDEX idx_cs_group ON chat_sessions(group_id);

CREATE TABLE session_participants (
    id               BIGSERIAL PRIMARY KEY,
    session_id       BIGINT       NOT NULL,
    user_id          BIGINT       NOT NULL,
    last_read_msg_id BIGINT       DEFAULT 0,
    muted            BOOLEAN      DEFAULT FALSE,
    joined_at        TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_sp_session FOREIGN KEY (session_id) REFERENCES chat_sessions(session_id) ON DELETE CASCADE,
    CONSTRAINT fk_sp_user    FOREIGN KEY (user_id)    REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT uq_sp_session_user UNIQUE (session_id, user_id)
);
CREATE INDEX idx_sp_user ON session_participants(user_id);
CREATE INDEX idx_sp_session ON session_participants(session_id);

CREATE TABLE files (
    file_id          BIGSERIAL PRIMARY KEY,
    uploader_id      BIGINT       NOT NULL,
    file_name        VARCHAR(255) NOT NULL,
    file_url         VARCHAR(500) NOT NULL,
    file_size        BIGINT       NOT NULL,
    file_type        VARCHAR(50),
    mime_type        VARCHAR(100),
    created_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_file_uploader FOREIGN KEY (uploader_id) REFERENCES users(user_id)
);
CREATE INDEX idx_file_uploader ON files(uploader_id);

CREATE TABLE messages (
    message_id       BIGSERIAL PRIMARY KEY,
    session_id       BIGINT       NOT NULL,
    sender_id        BIGINT       NOT NULL,
    message_type     SMALLINT     DEFAULT 1,
    content          TEXT,
    file_id          BIGINT,
    reply_to_msg_id  BIGINT,
    sent_at          TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_msg_session FOREIGN KEY (session_id) REFERENCES chat_sessions(session_id) ON DELETE CASCADE,
    CONSTRAINT fk_msg_sender  FOREIGN KEY (sender_id)  REFERENCES users(user_id),
    CONSTRAINT fk_msg_file    FOREIGN KEY (file_id)    REFERENCES files(file_id),
    CONSTRAINT chk_msg_type   CHECK (message_type IN (1, 2, 3, 4))
);
CREATE INDEX idx_msg_session_sent ON messages(session_id, sent_at DESC);
CREATE INDEX idx_msg_sender ON messages(sender_id);

CREATE TABLE message_status (
    id               BIGSERIAL PRIMARY KEY,
    message_id       BIGINT       NOT NULL,
    user_id          BIGINT       NOT NULL,
    read_status      SMALLINT     DEFAULT 0,
    read_at          TIMESTAMPTZ,
    CONSTRAINT fk_ms_msg  FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
    CONSTRAINT fk_ms_user FOREIGN KEY (user_id)    REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT chk_ms_status CHECK (read_status IN (0, 1)),
    CONSTRAINT uq_ms_msg_user UNIQUE (message_id, user_id)
);
CREATE INDEX idx_ms_user_status ON message_status(user_id, read_status);
CREATE INDEX idx_ms_msg ON message_status(message_id);

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated   BEFORE UPDATE ON users   FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_groups_updated  BEFORE UPDATE ON group_info FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_sessions_updated BEFORE UPDATE ON chat_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

- [ ] **Step 5: Create migrations/001_init_schema.down.sql**

```sql
DROP TABLE IF EXISTS message_status;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS session_participants;
DROP TABLE IF EXISTS chat_sessions;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS group_info;
DROP TABLE IF EXISTS friend_relations;
DROP TABLE IF EXISTS friend_groups;
DROP TABLE IF EXISTS users;
DROP FUNCTION IF EXISTS update_updated_at();
```

- [ ] **Step 6: Create docker-compose.yml, Dockerfile, Makefile**

As specified in the dev guide §4.2, §16.1, §16.2.

- [ ] **Step 7: Commit**

```bash
git add .
git commit -m "chore: scaffold project structure, migrations, and configs"
```

---

### Task 2: Utility Packages (errcode, response, jwt, hash)

**Covers:** §13, §12.1

**Files:**
- Create: `internal/pkg/errcode/errcode.go`
- Create: `internal/pkg/response/response.go`
- Create: `internal/pkg/jwt/jwt.go`
- Create: `internal/pkg/hash/hash.go`

**Interfaces:**
- Produces:
  - `errcode.ErrCode` type and constants (Success, InvalidParams, Unauthorized, etc.)
  - `response.OK(data)`, `response.BadRequest(msg)`, `response.Unauthorized(msg)`, `response.NotFound(msg)`, `response.Conflict(msg)`, `response.FromError(err)`
  - `response.APIError` struct (Code, Msg, HTTPStatus) implementing `error`
  - `jwt.Generate(userID int64, username, secret string, expiry time.Duration) (string, error)`
  - `jwt.Parse(tokenStr, secret string) (*Claims, error)` where Claims has UserID, Username
  - `hash.Bcrypt(password string) (string, error)`
  - `hash.VerifyBcrypt(password, hash string) bool`

- [ ] **Step 1: Create internal/pkg/errcode/errcode.go**

```go
package errcode

type ErrCode int

const (
	Success           ErrCode = 0
	InvalidParams     ErrCode = 40001
	Unauthorized      ErrCode = 40101
	Forbidden         ErrCode = 40301
	NotFound          ErrCode = 40401
	Conflict          ErrCode = 40901
	InternalError     ErrCode = 50001
	UsernameExists    ErrCode = 41001
	InvalidCredential ErrCode = 41002
	SessionNotFound   ErrCode = 41003
	NotParticipant    ErrCode = 41004
)
```

- [ ] **Step 2: Create internal/pkg/response/response.go**

```go
package response

import (
	"errors"
	"net/http"

	"chatgoo/internal/pkg/errcode"
)

type Response struct {
	Code    errcode.ErrCode `json:"code"`
	Message string          `json:"message"`
	Data    interface{}     `json:"data,omitempty"`
}

func OK(data interface{}) *Response {
	return &Response{Code: errcode.Success, Message: "success", Data: data}
}

type APIError struct {
	Code       errcode.ErrCode
	Msg        string
	HTTPStatus int
}

func (e *APIError) Error() string { return e.Msg }

func BadRequest(msg string) error {
	return &APIError{Code: errcode.InvalidParams, Msg: msg, HTTPStatus: http.StatusBadRequest}
}

func Unauthorized(msg string) error {
	return &APIError{Code: errcode.Unauthorized, Msg: msg, HTTPStatus: http.StatusUnauthorized}
}

func NotFound(msg string) error {
	return &APIError{Code: errcode.NotFound, Msg: msg, HTTPStatus: http.StatusNotFound}
}

func Conflict(msg string) error {
	return &APIError{Code: errcode.Conflict, Msg: msg, HTTPStatus: http.StatusConflict}
}

func FromError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	switch err.Error() {
	case "username already exists":
		return &APIError{Code: errcode.UsernameExists, Msg: err.Error(), HTTPStatus: http.StatusConflict}
	case "invalid username or password":
		return &APIError{Code: errcode.InvalidCredential, Msg: err.Error(), HTTPStatus: http.StatusUnauthorized}
	case "session not found":
		return &APIError{Code: errcode.SessionNotFound, Msg: err.Error(), HTTPStatus: http.StatusNotFound}
	case "user is not a participant of this session":
		return &APIError{Code: errcode.NotParticipant, Msg: err.Error(), HTTPStatus: http.StatusForbidden}
	default:
		return &APIError{Code: errcode.InternalError, Msg: "internal server error", HTTPStatus: http.StatusInternalServerError}
	}
}
```

- [ ] **Step 3: Create internal/pkg/jwt/jwt.go**

```go
package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Generate(userID int64, username, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Parse(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
```

- [ ] **Step 4: Create internal/pkg/hash/hash.go**

```go
package hash

import "golang.org/x/crypto/bcrypt"

func Bcrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func VerifyBcrypt(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
```

- [ ] **Step 5: Run go mod tidy and verify compilation**

```bash
cd E:\Go\chatgoo
go mod tidy
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/pkg/
git commit -m "feat: add utility packages (errcode, response, jwt, hash)"
```

---

### Task 3: Data Models

**Covers:** §7

**Files:**
- Create: `internal/model/user.go`
- Create: `internal/model/friend.go`
- Create: `internal/model/group.go`
- Create: `internal/model/session.go`
- Create: `internal/model/message.go`
- Create: `internal/model/file.go`

**Interfaces:**
- Produces: All Go structs with `json` and `db` tags matching the database schema
  - `model.User`, `model.FriendGroup`, `model.FriendRelation`
  - `model.GroupInfo`, `model.GroupMember`
  - `model.ChatSession`, `model.SessionParticipant`
  - `model.Message`, `model.MessageStatus`, `model.File`
  - All enum constants (UserStatus*, Gender*, FriendStatus*, GroupRole*, SessionType*, MessageType*, ReadStatus*)

- [ ] **Step 1: Create all model files**

Create each file as specified in §7.1–7.5. All structs use `json` and `db` tags. PasswordHash has `json:"-"`.

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/model/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/model/
git commit -m "feat: add data models for all entities"
```

---

### Task 4: Database Helper & User Repository

**Covers:** §8.1, §8.3

**Files:**
- Create: `internal/repository/db.go`
- Create: `internal/repository/user_repo.go`

**Interfaces:**
- Produces:
  - `repository.WithGoFrContext(ctx, gofrCtx) context.Context`
  - `repository.getDB(ctx) *sql.DB` (unexported helper)
  - `repository.UserRepository` interface: Create, GetByID, GetByUsername, Update, UpdateLastLogin, UpdateStatus, ListByIDs
  - `repository.NewUserRepository() UserRepository`

- [ ] **Step 1: Create internal/repository/db.go**

```go
package repository

import (
	"context"
	"database/sql"

	"gofr.dev/pkg/gofr"
)

type ctxKey struct{}

func WithGoFrContext(ctx context.Context, gofrCtx *gofr.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, gofrCtx)
}

func getDB(ctx context.Context) *sql.DB {
	gofrCtx, ok := ctx.Value(ctxKey{}).(*gofr.Context)
	if !ok || gofrCtx == nil {
		return nil
	}
	return gofrCtx.DB
}
```

- [ ] **Step 2: Create internal/repository/user_repo.go**

Implement full UserRepository per §8.1 with all 7 methods. Each method uses `getDB(ctx)` and parameterized SQL.

- [ ] **Step 3: Verify compilation**

```bash
go build ./internal/repository/...
```

- [ ] **Step 4: Commit**

```bash
git add internal/repository/
git commit -m "feat: add db helper and user repository"
```

---

### Task 5: WebSocket Module

**Covers:** §11

**Files:**
- Create: `internal/ws/hub.go`
- Create: `internal/ws/client.go`
- Create: `internal/ws/handler.go`

**Interfaces:**
- Produces:
  - `ws.WSMessage` struct (Type string, Data interface{})
  - `ws.Hub` struct with: NewHub(), Run(), RegisterSession(), UnregisterSession(), BroadcastToSession(), SendToUser(), IsOnline()
  - `ws.Client` struct with: NewClient(), ReadPump(), WritePump()
  - `ws.MessageHandler` interface: Handle(client *Client, message []byte)
  - `ws.NewDefaultMessageHandler(hub *Hub, gofrCtx *gofr.Context) MessageHandler`

- [ ] **Step 1: Create internal/ws/hub.go**

Per §11.1 — Hub manages client connections (userID → connections map) and session routing (sessionID → userIDs map). Channels for register/unregister/broadcast. `Run()` loop processes all three.

- [ ] **Step 2: Create internal/ws/client.go**

Per §11.2 — Client wraps gorilla/websocket conn. ReadPump reads messages and dispatches to handler. WritePump sends from channel with ping ticker. Constants: writeWait=10s, pongWait=60s, pingPeriod=30s, maxMessageSize=4096.

- [ ] **Step 3: Create internal/ws/handler.go**

Define `MessageHandler` interface and `DefaultMessageHandler` implementation. Handles incoming WS messages (ping, subscribe_session, etc.).

- [ ] **Step 4: Verify compilation**

```bash
go build ./internal/ws/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/ws/
git commit -m "feat: add WebSocket hub, client, and message handler"
```

---

### Task 6: User Management Module (Service + Handler + Middleware + Router + Main)

**Covers:** §9.1, §10, §12

**Files:**
- Create: `internal/service/user_service.go`
- Create: `internal/handler/user_handler.go`
- Create: `internal/handler/ws_handler.go`
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/cors.go`
- Create: `internal/router/router.go`
- Create: `cmd/server/main.go`

**Interfaces:**
- Consumes: UserRepository, jwt.Generate/Parse, hash.Bcrypt/VerifyBcrypt, response.*, errcode.*, ws.Hub
- Produces:
  - `service.UserService` interface: Register, Login, GetProfile, UpdateProfile, SearchUser
  - `service.NewUserService(userRepo, jwtSecret, jwtExpH)`
  - `handler.Register()`, `handler.Login()`, `handler.GetProfile()`, `handler.UpdateProfile()`, `handler.SearchUser()`
  - `handler.WSHandler(hub)` — WebSocket upgrade handler
  - `middleware.Auth()` — JWT auth middleware
  - `middleware.CORS()` — CORS middleware
  - `router.Register(app, hub)` — route registration
  - `main()` — application entry point

- [ ] **Step 1: Create internal/service/user_service.go**

Per §9.1 — Register (validate, check uniqueness, hash password, create user, generate JWT), Login (find user, verify password, update last_login, generate JWT), GetProfile, UpdateProfile, SearchUser.

- [ ] **Step 2: Create internal/handler/user_handler.go**

Per §10.3 — Handlers extract userID from context (set by auth middleware), bind request body, call service, return response.OK or response.FromError.

- [ ] **Step 3: Create internal/handler/ws_handler.go**

Per §11.3 — WSHandler upgrades HTTP to WebSocket, creates Client, registers with hub, starts ReadPump/WritePump goroutines.

- [ ] **Step 4: Create internal/middleware/auth.go**

Per §12.1 — Extract Bearer token from Authorization header, parse JWT, inject userID/username into context via `c.Set()`.

- [ ] **Step 5: Create internal/middleware/cors.go**

Per §12.2 — Set CORS headers, handle OPTIONS preflight.

- [ ] **Step 6: Create internal/router/router.go**

Per §10.2 — Register all routes. Auth routes (/auth/register, /auth/login) are public. All other routes use `middleware.Auth`.

- [ ] **Step 7: Create cmd/server/main.go**

Per §10.1 — Initialize GoFr app, create Hub (go hub.Run()), create repositories, create services, register dependencies via `app.AddDependency()`, register middleware, register routes, call `app.Run()`.

- [ ] **Step 8: Verify full build**

```bash
cd E:\Go\chatgoo
go mod tidy
go build ./...
```

- [ ] **Step 9: Commit**

```bash
git add internal/service/ internal/handler/ internal/middleware/ internal/router/ cmd/
git commit -m "feat: add user management module with auth, handlers, and router"
```

---

### Task 7: AGENTS.md & README.md

**Covers:** Project documentation

**Files:**
- Create: `AGENTS.md`
- Create: `README.md`

- [ ] **Step 1: Create AGENTS.md**

Document the project structure, coding conventions, architecture decisions, and guidance for AI agents working on this codebase. Include:
- Project overview and tech stack
- Directory structure explanation
- Layer responsibilities (Handler/Service/Repository)
- How to add new features (new entity workflow)
- Testing conventions
- Key patterns (GoFr context, dependency injection, error handling)

- [ ] **Step 2: Create README.md**

Standard project README:
- Project title and description
- Tech stack
- Quick start (docker-compose up, go run)
- API overview
- Configuration reference
- Development guide

- [ ] **Step 3: Commit**

```bash
git add AGENTS.md README.md
git commit -m "docs: add AGENTS.md and README.md"
```
