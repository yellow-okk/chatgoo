# 在线聊天应用后端开发文档（GoFr + PostgreSQL）

> **目标读者**：AI Coding Agent / 后端开发工程师
> **技术栈**：Go 1.21+ · GoFr v1.x · PostgreSQL 15+ · Redis 7（可选）
> **项目代号**：`chat-app`
> **版本**：v1.0
> **最后更新**：2026-07-04

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术栈说明](#2-技术栈说明)
3. [项目目录结构](#3-项目目录结构)
4. [环境准备](#4-环境准备)
5. [数据库设计与初始化](#5-数据库设计与初始化)
6. [配置管理](#6-配置管理)
7. [数据模型（Go Struct）](#7-数据模型go-struct)
8. [数据访问层（Repository）](#8-数据访问层repository)
9. [业务逻辑层（Service）](#9-业务逻辑层service)
10. [HTTP 路由与 Handler](#10-http-路由与-handler)
11. [WebSocket 实时通信](#11-websocket-实时通信)
12. [中间件](#12-中间件)
13. [错误处理与统一响应](#13-错误处理与统一响应)
14. [API 接口清单](#14-api-接口清单)
15. [测试](#15-测试)
16. [部署](#16-部署)
17. [开发约定与注意事项](#17-开发约定与注意事项)

---

## 1. 项目概述

### 1.1 项目目标

构建一个对标 QQ（原型级别）的在线聊天应用后端，提供以下核心能力：

- 用户注册、登录、个人资料管理
- 好友关系管理（申请、同意、删除、分组）
- 一对一私聊与多人群聊
- 实时消息推送（WebSocket）
- 离线消息存储与已读回执
- 文件/图片消息发送

### 1.2 设计原则

| 原则 | 说明 |
|------|------|
| 分层架构 | Handler → Service → Repository 三层清晰分离 |
| 接口优先 | 所有业务逻辑通过接口定义，便于 Mock 测试 |
| 配置外置 | 数据库、Redis、JWT 等配置全部走环境变量 |
| 幂等性 | 消息发送、好友申请等关键操作支持幂等 |
| 可观测 | GoFr 内置 tracing/metrics，关键路径加日志 |

---

## 2. 技术栈说明

### 2.1 核心依赖

| 组件 | 版本 | 用途 |
|------|------|------|
| Go | 1.21+ | 编程语言 |
| [GoFr](https://gofr.dev/) | v1.x | HTTP 框架、依赖注入、配置管理 |
| PostgreSQL | 15+ | 主数据库 |
| pgx | v5 | PostgreSQL 驱动（GoFr 默认） |
| Redis | 7（可选） | 会话缓存、未读消息计数、限流 |
| gorilla/websocket | v1.5 | WebSocket 实现 |
| golang-jwt/jwt | v5 | JWT 鉴权 |
| bcrypt | - | 密码哈希 |
| uuid | - | UUID 生成 |
| testify | v1.8+ | 单元测试 |

### 2.2 GoFr 框架特性说明

GoFr 是一个 Go 语言的微服务框架，核心特性：

- **自动配置加载**：从 `.env` / 环境变量读取配置，注入 `ctx.Config("KEY")`
- **依赖注入**：通过 `app.AddDependency()` 注册服务，Handler 中通过 `ctx.Get("serviceName")` 获取
- **内置数据库连接**：自动创建 `*sql.DB`，通过 `ctx.DB()` 访问
- **内置 Redis**：配置 `REDIS_HOST` 后通过 `ctx.Redis()` 访问
- **结构化日志**：`ctx.Logger.Infof(...)`
- **OpenAPI 自动生成**：路由自动生成 Swagger 文档
- **请求/响应绑定**：`ctx.Bind(&req)` / `ctx.Bind(&resp)`

---

## 3. 项目目录结构

```
chat-app/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置定义
│   ├── model/
│   │   ├── user.go
│   │   ├── friend.go
│   │   ├── group.go
│   │   ├── session.go
│   │   ├── message.go
│   │   └── file.go
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── friend_repo.go
│   │   ├── group_repo.go
│   │   ├── session_repo.go
│   │   ├── message_repo.go
│   │   └── file_repo.go
│   ├── service/
│   │   ├── user_service.go
│   │   ├── friend_service.go
│   │   ├── group_service.go
│   │   ├── message_service.go
│   │   └── file_service.go
│   ├── handler/
│   │   ├── user_handler.go
│   │   ├── friend_handler.go
│   │   ├── group_handler.go
│   │   ├── message_handler.go
│   │   └── file_handler.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── cors.go
│   ├── ws/
│   │   ├── hub.go               # WebSocket 连接管理
│   │   ├── client.go            # 单个连接封装
│   │   └── handler.go           # 消息分发
│   ├── router/
│   │   └── router.go            # 路由注册
│   └── pkg/
│       ├── jwt/
│       │   └── jwt.go
│       ├── hash/
│       │   └── hash.go
│       ├── response/
│       │   └── response.go      # 统一响应封装
│       └── errcode/
│           └── errcode.go       # 错误码定义
├── migrations/
│   ├── 001_init_schema.up.sql
│   └── 001_init_schema.down.sql
├── configs/
│   └── .env.example
├── scripts/
│   └── seed.sql                 # 测试数据
├── docs/
│   └── api.md
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## 4. 环境准备

### 4.1 本地开发环境

```bash
# Go 1.21+
go version

# PostgreSQL 15+
psql --version

# Redis 7+（可选）
redis-cli --version
```

### 4.2 Docker Compose 一键启动

`docker-compose.yml`：

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: chat-postgres
    environment:
      POSTGRES_USER: chat
      POSTGRES_PASSWORD: chat123
      POSTGRES_DB: chatapp
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./migrations/001_init_schema.up.sql:/docker-entrypoint-initdb.d/01-schema.sql
      - ./scripts/seed.sql:/docker-entrypoint-initdb.d/02-seed.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U chat -d chatapp"]
      interval: 5s
      timeout: 3s
      retries: 10

  redis:
    image: redis:7-alpine
    container_name: chat-redis
    ports:
      - "6379:6379"

  app:
    build: .
    container_name: chat-app
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    environment:
      - APP_ENV=development
      - HTTP_PORT=8000
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=chat
      - DB_PASSWORD=chat123
      - DB_NAME=chatapp
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=your-super-secret-key-change-in-prod
      - JWT_EXPIRE_HOURS=72
    ports:
      - "8000:8000"

volumes:
  pgdata:
```

### 4.3 启动命令

```bash
# 启动所有依赖
docker-compose up -d postgres redis

# 本地运行应用（热重载）
go install github.com/cosmtrek/air@latest
air

# 或直接运行
go run cmd/server/main.go
```

---

## 5. 数据库设计与初始化

### 5.1 完整建表 SQL

文件路径：`migrations/001_init_schema.up.sql`

```sql
-- ============================================================
-- 在线聊天应用 - 数据库初始化脚本
-- Database: PostgreSQL 15+
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- 1. 用户表
-- ============================================================
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

-- ============================================================
-- 2. 好友分组表
-- ============================================================
CREATE TABLE friend_groups (
    group_id         BIGSERIAL PRIMARY KEY,
    user_id          BIGINT       NOT NULL,
    group_name       VARCHAR(50)  NOT NULL,
    created_at       TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_fg_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT uq_fg_user_group UNIQUE (user_id, group_name)
);
CREATE INDEX idx_fg_user ON friend_groups(user_id);

-- ============================================================
-- 3. 好友关系表
-- ============================================================
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

-- ============================================================
-- 4. 群组表
-- ============================================================
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

-- ============================================================
-- 5. 群成员表
-- ============================================================
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

-- ============================================================
-- 6. 聊天会话表
-- ============================================================
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

-- ============================================================
-- 7. 会话参与者表
-- ============================================================
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

-- ============================================================
-- 8. 消息表
-- ============================================================
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

-- ============================================================
-- 9. 文件表（注意：messages 引用 files，需先建）
-- ============================================================
-- 注意：实际建表顺序中 files 应在 messages 之前。
-- 此处为逻辑分组展示，迁移脚本中 files 表定义需置于 messages 之前。

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

-- ============================================================
-- 10. 消息状态表（已读回执）
-- ============================================================
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

-- ============================================================
-- 触发器：自动更新 updated_at
-- ============================================================
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

> ⚠️ **建表顺序注意**：实际迁移脚本中，`files` 表必须在 `messages` 表之前创建（因 `messages.file_id` 外键引用 `files.file_id`）。请将 `files` 表定义移至 `messages` 之前。

### 5.2 回滚脚本

`migrations/001_init_schema.down.sql`：

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

### 5.3 迁移工具推荐

使用 [`golang-migrate`](https://github.com/golang-migrate/migrate)：

```bash
# 安装
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 执行迁移
migrate -path migrations -database "postgres://chat:chat123@localhost:5432/chatapp?sslmode=disable" up

# 回滚
migrate -path migrations -database "postgres://chat:chat123@localhost:5432/chatapp?sslmode=disable" down 1
```

---

## 6. 配置管理

### 6.1 环境变量示例

`configs/.env.example`：

```env
# 应用配置
APP_ENV=development
APP_NAME=chat-app
HTTP_PORT=8000

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=chat
DB_PASSWORD=chat123
DB_NAME=chatapp
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300s

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-super-secret-key-change-in-prod
JWT_EXPIRE_HOURS=72

# 文件上传
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=50

# WebSocket
WS_READ_BUFFER_SIZE=10240
WS_WRITE_BUFFER_SIZE=10240
WS_PING_INTERVAL=30s
```

### 6.2 GoFr 自动加载

GoFr 会自动读取 `.env` 文件和环境变量，无需手动解析。在代码中通过 `ctx.Config("KEY")` 获取。

---

## 7. 数据模型（Go Struct）

### 7.1 用户模型

`internal/model/user.go`：

```go
package model

import "time"

type User struct {
    UserID       int64      `json:"user_id" db:"user_id"`
    Username     string     `json:"username" db:"username"`
    PasswordHash string     `json:"-" db:"password_hash"`
    Nickname     string     `json:"nickname" db:"nickname"`
    AvatarURL    string     `json:"avatar_url" db:"avatar_url"`
    Gender       int16      `json:"gender" db:"gender"`
    Signature    string     `json:"signature" db:"signature"`
    Region       string     `json:"region" db:"region"`
    Birthday     *time.Time `json:"birthday" db:"birthday"`
    Status       int16      `json:"status" db:"status"`
    LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
    CreatedAt    time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// UserStatus 枚举
const (
    UserStatusOffline  int16 = 0
    UserStatusOnline   int16 = 1
    UserStatusAway     int16 = 2
)

// Gender 枚举
const (
    GenderUnknown int16 = 0
    GenderMale    int16 = 1
    GenderFemale  int16 = 2
)
```

### 7.2 好友关系模型

`internal/model/friend.go`：

```go
package model

import "time"

type FriendGroup struct {
    GroupID   int64     `json:"group_id" db:"group_id"`
    UserID    int64     `json:"user_id" db:"user_id"`
    GroupName string    `json:"group_name" db:"group_name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type FriendRelation struct {
    RelationID int64      `json:"relation_id" db:"relation_id"`
    UserID     int64      `json:"user_id" db:"user_id"`
    FriendID   int64      `json:"friend_id" db:"friend_id"`
    GroupID    *int64     `json:"group_id" db:"group_id"`
    Remark     string     `json:"remark" db:"remark"`
    Status     int16      `json:"status" db:"status"`
    AppliedAt  time.Time  `json:"applied_at" db:"applied_at"`
    ApprovedAt *time.Time `json:"approved_at" db:"approved_at"`
    CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// FriendStatus 枚举
const (
    FriendStatusPending  int16 = 0 // 待处理
    FriendStatusAccepted int16 = 1 // 已接受
    FriendStatusRejected int16 = 2 // 已拒绝
    FriendStatusBlocked  int16 = 3 // 已拉黑
)
```

### 7.3 群组模型

`internal/model/group.go`：

```go
package model

import "time"

type GroupInfo struct {
    GroupID      int64     `json:"group_id" db:"group_id"`
    GroupName    string    `json:"group_name" db:"group_name"`
    OwnerID      int64     `json:"owner_id" db:"owner_id"`
    AvatarURL    string    `json:"avatar_url" db:"avatar_url"`
    Announcement string    `json:"announcement" db:"announcement"`
    MaxMembers   int       `json:"max_members" db:"max_members"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type GroupMember struct {
    ID      int64     `json:"id" db:"id"`
    GroupID int64     `json:"group_id" db:"group_id"`
    UserID  int64     `json:"user_id" db:"user_id"`
    Role    int16     `json:"role" db:"role"`
    JoinAt  time.Time `json:"join_at" db:"join_at"`
}

// GroupRole 枚举
const (
    GroupRoleMember int16 = 0
    GroupRoleAdmin  int16 = 1
    GroupRoleOwner  int16 = 2
)
```

### 7.4 会话与消息模型

`internal/model/session.go`：

```go
package model

import "time"

type ChatSession struct {
    SessionID     int64      `json:"session_id" db:"session_id"`
    SessionType   int16      `json:"session_type" db:"session_type"`
    TargetUserID  *int64     `json:"target_user_id" db:"target_user_id"`
    GroupID       *int64     `json:"group_id" db:"group_id"`
    LastMessageID *int64     `json:"last_message_id" db:"last_message_id"`
    LastMessageAt *time.Time `json:"last_message_at" db:"last_message_at"`
    CreatedAt     time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type SessionParticipant struct {
    ID             int64     `json:"id" db:"id"`
    SessionID      int64     `json:"session_id" db:"session_id"`
    UserID         int64     `json:"user_id" db:"user_id"`
    LastReadMsgID  int64     `json:"last_read_msg_id" db:"last_read_msg_id"`
    Muted          bool      `json:"muted" db:"muted"`
    JoinedAt       time.Time `json:"joined_at" db:"joined_at"`
}

// SessionType 枚举
const (
    SessionTypePrivate int16 = 1
    SessionTypeGroup   int16 = 2
)
```

`internal/model/message.go`：

```go
package model

import "time"

type Message struct {
    MessageID      int64     `json:"message_id" db:"message_id"`
    SessionID      int64     `json:"session_id" db:"session_id"`
    SenderID       int64     `json:"sender_id" db:"sender_id"`
    MessageType    int16     `json:"message_type" db:"message_type"`
    Content        string    `json:"content" db:"content"`
    FileID         *int64    `json:"file_id" db:"file_id"`
    ReplyToMsgID   *int64    `json:"reply_to_msg_id" db:"reply_to_msg_id"`
    SentAt         time.Time `json:"sent_at" db:"sent_at"`
}

type MessageStatus struct {
    ID         int64      `json:"id" db:"id"`
    MessageID  int64      `json:"message_id" db:"message_id"`
    UserID     int64      `json:"user_id" db:"user_id"`
    ReadStatus int16      `json:"read_status" db:"read_status"`
    ReadAt     *time.Time `json:"read_at" db:"read_at"`
}

// MessageType 枚举
const (
    MessageTypeText  int16 = 1
    MessageTypeImage int16 = 2
    MessageTypeFile  int16 = 3
    MessageTypeVoice int16 = 4
)

// ReadStatus 枚举
const (
    ReadStatusUnread int16 = 0
    ReadStatusRead   int16 = 1
)
```

### 7.5 文件模型

`internal/model/file.go`：

```go
package model

import "time"

type File struct {
    FileID     int64     `json:"file_id" db:"file_id"`
    UploaderID int64     `json:"uploader_id" db:"uploader_id"`
    FileName   string    `json:"file_name" db:"file_name"`
    FileURL    string    `json:"file_url" db:"file_url"`
    FileSize   int64     `json:"file_size" db:"file_size"`
    FileType   string    `json:"file_type" db:"file_type"`
    MimeType   string    `json:"mime_type" db:"mime_type"`
    CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
```

---

## 8. 数据访问层（Repository）

### 8.1 接口定义

`internal/repository/user_repo.go`：

```go
package repository

import (
    "context"
    "chat-app/internal/model"
)

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, userID int64) (*model.User, error)
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    UpdateLastLogin(ctx context.Context, userID int64) error
    UpdateStatus(ctx context.Context, userID int64, status int16) error
    ListByIDs(ctx context.Context, ids []int64) ([]*model.User, error)
}

type userRepository struct{}

func NewUserRepository() UserRepository {
    return &userRepository{}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    db := gofr.ContextFrom(ctx).DB  // GoFr 注入的 *sql.DB
    query := `
        INSERT INTO users (username, password_hash, nickname, avatar_url, gender, signature, region, birthday)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING user_id, created_at, updated_at
    `
    return db.QueryRowContext(ctx, query,
        user.Username, user.PasswordHash, user.Nickname,
        user.AvatarURL, user.Gender, user.Signature,
        user.Region, user.Birthday,
    ).Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByID(ctx context.Context, userID int64) (*model.User, error) {
    db := getDB(ctx)
    var u model.User
    query := `
        SELECT user_id, username, password_hash, nickname, avatar_url,
               gender, signature, region, birthday, status, last_login_at,
               created_at, updated_at
        FROM users WHERE user_id = $1
    `
    err := db.QueryRowContext(ctx, query, userID).Scan(
        &u.UserID, &u.Username, &u.PasswordHash, &u.Nickname, &u.AvatarURL,
        &u.Gender, &u.Signature, &u.Region, &u.Birthday, &u.Status, &u.LastLoginAt,
        &u.CreatedAt, &u.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    db := getDB(ctx)
    var u model.User
    query := `
        SELECT user_id, username, password_hash, nickname, avatar_url,
               gender, signature, region, birthday, status, last_login_at,
               created_at, updated_at
        FROM users WHERE username = $1
    `
    err := db.QueryRowContext(ctx, query, username).Scan(
        &u.UserID, &u.Username, &u.PasswordHash, &u.Nickname, &u.AvatarURL,
        &u.Gender, &u.Signature, &u.Region, &u.Birthday, &u.Status, &u.LastLoginAt,
        &u.CreatedAt, &u.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &u, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
    db := getDB(ctx)
    query := `
        UPDATE users
        SET nickname = $1, avatar_url = $2, gender = $3,
            signature = $4, region = $5, birthday = $6
        WHERE user_id = $7
    `
    _, err := db.ExecContext(ctx, query,
        user.Nickname, user.AvatarURL, user.Gender,
        user.Signature, user.Region, user.Birthday, user.UserID,
    )
    return err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
    db := getDB(ctx)
    _, err := db.ExecContext(ctx,
        `UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE user_id = $1`,
        userID,
    )
    return err
}

func (r *userRepository) UpdateStatus(ctx context.Context, userID int64, status int16) error {
    db := getDB(ctx)
    _, err := db.ExecContext(ctx,
        `UPDATE users SET status = $1 WHERE user_id = $2`,
        status, userID,
    )
    return err
}

func (r *userRepository) ListByIDs(ctx context.Context, ids []int64) ([]*model.User, error) {
    if len(ids) == 0 {
        return nil, nil
    }
    db := getDB(ctx)
    query := `
        SELECT user_id, username, nickname, avatar_url, gender, signature, region, status
        FROM users WHERE user_id = ANY($1)
    `
    rows, err := db.QueryContext(ctx, query, ids)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*model.User
    for rows.Next() {
        var u model.User
        if err := rows.Scan(
            &u.UserID, &u.Username, &u.Nickname, &u.AvatarURL,
            &u.Gender, &u.Signature, &u.Region, &u.Status,
        ); err != nil {
            return nil, err
        }
        users = append(users, &u)
    }
    return users, nil
}
```

> **说明**：`getDB(ctx)` 是一个辅助函数，从 GoFr 的 context 中提取 `*sql.DB`。具体实现见 8.3 节。

### 8.2 消息 Repository（含事务示例）

`internal/repository/message_repo.go`：

```go
package repository

import (
    "context"
    "chat-app/internal/model"
)

type MessageRepository interface {
    Create(ctx context.Context, msg *model.Message) error
    GetByID(ctx context.Context, msgID int64) (*model.Message, error)
    ListBySession(ctx context.Context, sessionID int64, beforeID int64, limit int) ([]*model.Message, error)
    UpdateReadStatus(ctx context.Context, msgID, userID int64) error
    GetUnreadCount(ctx context.Context, userID int64) (map[int64]int, error)
}

type messageRepository struct{}

func NewMessageRepository() MessageRepository {
    return &messageRepository{}
}

// Create 发送消息（事务：插入消息 + 更新会话 last_message + 初始化已读状态）
func (r *messageRepository) Create(ctx context.Context, msg *model.Message) error {
    db := getDB(ctx)
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. 插入消息
    query := `
        INSERT INTO messages (session_id, sender_id, message_type, content, file_id, reply_to_msg_id)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING message_id, sent_at
    `
    err = tx.QueryRowContext(ctx, query,
        msg.SessionID, msg.SenderID, msg.MessageType,
        msg.Content, msg.FileID, msg.ReplyToMsgID,
    ).Scan(&msg.MessageID, &msg.SentAt)
    if err != nil {
        return err
    }

    // 2. 更新会话最后消息
    _, err = tx.ExecContext(ctx, `
        UPDATE chat_sessions
        SET last_message_id = $1, last_message_at = $2
        WHERE session_id = $3
    `, msg.MessageID, msg.SentAt, msg.SessionID)
    if err != nil {
        return err
    }

    // 3. 为发送者标记已读
    _, err = tx.ExecContext(ctx, `
        INSERT INTO message_status (message_id, user_id, read_status, read_at)
        VALUES ($1, $2, 1, CURRENT_TIMESTAMP)
        ON CONFLICT (message_id, user_id) DO NOTHING
    `, msg.MessageID, msg.SenderID)
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (r *messageRepository) ListBySession(ctx context.Context, sessionID int64, beforeID int64, limit int) ([]*model.Message, error) {
    db := getDB(ctx)
    if limit <= 0 || limit > 100 {
        limit = 50
    }

    var query string
    var args []interface{}
    if beforeID > 0 {
        query = `
            SELECT message_id, session_id, sender_id, message_type, content,
                   file_id, reply_to_msg_id, sent_at
            FROM messages
            WHERE session_id = $1 AND message_id < $2
            ORDER BY message_id DESC
            LIMIT $3
        `
        args = []interface{}{sessionID, beforeID, limit}
    } else {
        query = `
            SELECT message_id, session_id, sender_id, message_type, content,
                   file_id, reply_to_msg_id, sent_at
            FROM messages
            WHERE session_id = $1
            ORDER BY message_id DESC
            LIMIT $2
        `
        args = []interface{}{sessionID, limit}
    }

    rows, err := db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var messages []*model.Message
    for rows.Next() {
        var m model.Message
        if err := rows.Scan(
            &m.MessageID, &m.SessionID, &m.SenderID, &m.MessageType,
            &m.Content, &m.FileID, &m.ReplyToMsgID, &m.SentAt,
        ); err != nil {
            return nil, err
        }
        messages = append(messages, &m)
    }
    return messages, nil
}

// GetUnreadCount 获取用户所有会话的未读消息数
// 返回 map[session_id]unread_count
func (r *messageRepository) GetUnreadCount(ctx context.Context, userID int64) (map[int64]int, error) {
    db := getDB(ctx)
    query := `
        SELECT sp.session_id, COUNT(m.message_id) AS unread
        FROM session_participants sp
        JOIN messages m ON m.session_id = sp.session_id
        LEFT JOIN message_status ms ON ms.message_id = m.message_id AND ms.user_id = sp.user_id
        WHERE sp.user_id = $1
          AND m.message_id > sp.last_read_msg_id
          AND (ms.read_status IS NULL OR ms.read_status = 0)
          AND m.sender_id <> $1
        GROUP BY sp.session_id
    `
    rows, err := db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make(map[int64]int)
    for rows.Next() {
        var sessionID int64
        var count int
        if err := rows.Scan(&sessionID, &count); err != nil {
            return nil, err
        }
        result[sessionID] = count
    }
    return result, nil
}
```

### 8.3 DB 辅助函数

`internal/repository/db.go`：

```go
package repository

import (
    "context"
    "database/sql"

    "gofr.dev/pkg/gofr"
)

// GoFrContextKey 用于在 context 中传递 GoFr 上下文
type ctxKey struct{}

// WithGoFrContext 将 GoFr 上下文注入到标准 context 中
func WithGoFrContext(ctx context.Context, gofrCtx *gofr.Context) context.Context {
    return context.WithValue(ctx, ctxKey{}, gofrCtx)
}

// getDB 从 context 中提取 *sql.DB
func getDB(ctx context.Context) *sql.DB {
    gofrCtx, ok := ctx.Value(ctxKey{}).(*gofr.Context)
    if !ok || gofrCtx == nil {
        return nil
    }
    return gofrCtx.DB
}
```

### 8.4 其他 Repository 接口

为节省篇幅，仅列出接口定义，实现模式与 `UserRepository` 一致。

```go
// internal/repository/friend_repo.go
type FriendRepository interface {
    CreateGroup(ctx context.Context, g *model.FriendGroup) error
    ListGroups(ctx context.Context, userID int64) ([]*model.FriendGroup, error)
    DeleteGroup(ctx context.Context, groupID, userID int64) error

    ApplyFriend(ctx context.Context, fr *model.FriendRelation) error
    ApproveFriend(ctx context.Context, relationID int64) error
    RejectFriend(ctx context.Context, relationID int64) error
    RemoveFriend(ctx context.Context, userID, friendID int64) error
    ListFriends(ctx context.Context, userID int64) ([]*model.FriendRelation, error)
    ListPendingRequests(ctx context.Context, userID int64) ([]*model.FriendRelation, error)
    GetRelation(ctx context.Context, userID, friendID int64) (*model.FriendRelation, error)
}

// internal/repository/group_repo.go
type GroupRepository interface {
    Create(ctx context.Context, g *model.GroupInfo, ownerID int64) (int64, error)
    GetByID(ctx context.Context, groupID int64) (*model.GroupInfo, error)
    Update(ctx context.Context, g *model.GroupInfo) error
    Dismiss(ctx context.Context, groupID, ownerID int64) error

    AddMember(ctx context.Context, groupID, userID int64, role int16) error
    RemoveMember(ctx context.Context, groupID, userID int64) error
    UpdateMemberRole(ctx context.Context, groupID, userID int64, role int16) error
    ListMembers(ctx context.Context, groupID int64) ([]*model.GroupMember, error)
    IsMember(ctx context.Context, groupID, userID int64) (bool, error)
}

// internal/repository/session_repo.go
type SessionRepository interface {
    GetOrCreatePrivateSession(ctx context.Context, userA, userB int64) (int64, error)
    GetOrCreateGroupSession(ctx context.Context, groupID int64) (int64, error)
    GetByID(ctx context.Context, sessionID int64) (*model.ChatSession, error)
    ListByUser(ctx context.Context, userID int64) ([]*model.ChatSession, error)
    UpdateLastRead(ctx context.Context, sessionID, userID, msgID int64) error
    Mute(ctx context.Context, sessionID, userID int64, muted bool) error
}

// internal/repository/file_repo.go
type FileRepository interface {
    Create(ctx context.Context, f *model.File) error
    GetByID(ctx context.Context, fileID int64) (*model.File, error)
}
```

---

## 9. 业务逻辑层（Service）

### 9.1 用户服务

`internal/service/user_service.go`：

```go
package service

import (
    "context"
    "errors"
    "strings"
    "time"

    "chat-app/internal/model"
    "chat-app/internal/pkg/hash"
    "chat-app/internal/pkg/jwt"
    "chat-app/internal/repository"
)

var (
    ErrUsernameExists      = errors.New("username already exists")
    ErrInvalidCredentials  = errors.New("invalid username or password")
    ErrUsernameTooShort    = errors.New("username must be at least 3 characters")
    ErrPasswordTooShort    = errors.New("password must be at least 6 characters")
)

type UserService interface {
    Register(ctx context.Context, req *RegisterRequest) (*model.User, string, error)
    Login(ctx context.Context, username, password string) (*model.User, string, error)
    GetProfile(ctx context.Context, userID int64) (*model.User, error)
    UpdateProfile(ctx context.Context, user *model.User) error
    SearchUser(ctx context.Context, keyword string) ([]*model.User, error)
}

type RegisterRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Nickname string `json:"nickname"`
}

type userService struct {
    userRepo  repository.UserRepository
    jwtSecret string
    jwtExpH   int
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string, jwtExpH int) UserService {
    return &userService{
        userRepo:  userRepo,
        jwtSecret: jwtSecret,
        jwtExpH:   jwtExpH,
    }
}

func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*model.User, string, error) {
    // 参数校验
    if len(strings.TrimSpace(req.Username)) < 3 {
        return nil, "", ErrUsernameTooShort
    }
    if len(req.Password) < 6 {
        return nil, "", ErrPasswordTooShort
    }

    // 检查用户名是否已存在
    existing, err := s.userRepo.GetByUsername(ctx, req.Username)
    if err == nil && existing != nil {
        return nil, "", ErrUsernameExists
    }

    // 哈希密码
    hashed, err := hash.Bcrypt(req.Password)
    if err != nil {
        return nil, "", err
    }

    user := &model.User{
        Username:     req.Username,
        PasswordHash: hashed,
        Nickname:     req.Nickname,
        Status:       model.UserStatusOffline,
    }
    if user.Nickname == "" {
        user.Nickname = req.Username
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, "", err
    }

    // 生成 JWT
    token, err := jwt.Generate(user.UserID, user.Username, s.jwtSecret, time.Duration(s.jwtExpH)*time.Hour)
    if err != nil {
        return nil, "", err
    }

    return user, token, nil
}

func (s *userService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
    user, err := s.userRepo.GetByUsername(ctx, username)
    if err != nil || user == nil {
        return nil, "", ErrInvalidCredentials
    }

    if !hash.VerifyBcrypt(password, user.PasswordHash) {
        return nil, "", ErrInvalidCredentials
    }

    if err := s.userRepo.UpdateLastLogin(ctx, user.UserID); err != nil {
        return nil, "", err
    }

    token, err := jwt.Generate(user.UserID, user.Username, s.jwtSecret, time.Duration(s.jwtExpH)*time.Hour)
    if err != nil {
        return nil, "", err
    }

    return user, token, nil
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
    return s.userRepo.GetByID(ctx, userID)
}

func (s *userService) UpdateProfile(ctx context.Context, user *model.User) error {
    return s.userRepo.Update(ctx, user)
}

func (s *userService) SearchUser(ctx context.Context, keyword string) ([]*model.User, error) {
    // 实际实现需在 Repository 中添加搜索方法
    return nil, nil
}
```

### 9.2 消息服务

`internal/service/message_service.go`：

```go
package service

import (
    "context"
    "errors"
    "time"

    "chat-app/internal/model"
    "chat-app/internal/repository"
    "chat-app/internal/ws"
)

var (
    ErrSessionNotFound = errors.New("session not found")
    ErrNotParticipant  = errors.New("user is not a participant of this session")
    ErrEmptyMessage    = errors.New("message content cannot be empty")
)

type SendMessageRequest struct {
    SessionID    int64  `json:"session_id"`
    MessageType  int16  `json:"message_type"`
    Content      string `json:"content"`
    FileID       *int64 `json:"file_id"`
    ReplyToMsgID *int64 `json:"reply_to_msg_id"`
}

type MessageService interface {
    Send(ctx context.Context, senderID int64, req *SendMessageRequest) (*model.Message, error)
    History(ctx context.Context, userID, sessionID, beforeID int64, limit int) ([]*model.Message, error)
    MarkRead(ctx context.Context, userID, sessionID, msgID int64) error
    UnreadCount(ctx context.Context, userID int64) (map[int64]int, error)
}

type messageService struct {
    msgRepo     repository.MessageRepository
    sessionRepo repository.SessionRepository
    hub         *ws.Hub
}

func NewMessageService(
    msgRepo repository.MessageRepository,
    sessionRepo repository.SessionRepository,
    hub *ws.Hub,
) MessageService {
    return &messageService{
        msgRepo:     msgRepo,
        sessionRepo: sessionRepo,
        hub:         hub,
    }
}

func (s *messageService) Send(ctx context.Context, senderID int64, req *SendMessageRequest) (*model.Message, error) {
    // 1. 校验会话存在且发送者是参与者
    session, err := s.sessionRepo.GetByID(ctx, req.SessionID)
    if err != nil || session == nil {
        return nil, ErrSessionNotFound
    }

    // 2. 校验消息内容
    if req.MessageType == model.MessageTypeText && req.Content == "" {
        return nil, ErrEmptyMessage
    }

    // 3. 持久化消息
    msg := &model.Message{
        SessionID:    req.SessionID,
        SenderID:     senderID,
        MessageType:  req.MessageType,
        Content:      req.Content,
        FileID:       req.FileID,
        ReplyToMsgID: req.ReplyToMsgID,
    }
    if err := s.msgRepo.Create(ctx, msg); err != nil {
        return nil, err
    }

    // 4. 通过 WebSocket 推送给会话参与者
    s.hub.BroadcastToSession(req.SessionID, &ws.WSMessage{
        Type: "new_message",
        Data: msg,
    })

    return msg, nil
}

func (s *messageService) History(ctx context.Context, userID, sessionID, beforeID int64, limit int) ([]*model.Message, error) {
    // TODO: 校验 userID 是否是 session 的参与者
    return s.msgRepo.ListBySession(ctx, sessionID, beforeID, limit)
}

func (s *messageService) MarkRead(ctx context.Context, userID, sessionID, msgID int64) error {
    return s.sessionRepo.UpdateLastRead(ctx, sessionID, userID, msgID)
}

func (s *messageService) UnreadCount(ctx context.Context, userID int64) (map[int64]int, error) {
    return s.msgRepo.GetUnreadCount(ctx, userID)
}
```

---

## 10. HTTP 路由与 Handler

### 10.1 主入口

`cmd/server/main.go`：

```go
package main

import (
    "chat-app/internal/middleware"
    "chat-app/internal/repository"
    "chat-app/internal/router"
    "chat-app/internal/service"
    "chat-app/internal/ws"

    "gofr.dev/pkg/gofr"
)

func main() {
    app := gofr.New()

    // 初始化 WebSocket Hub
    hub := ws.NewHub()
    go hub.Run()

    // 初始化 Repository
    userRepo := repository.NewUserRepository()
    friendRepo := repository.NewFriendRepository()
    groupRepo := repository.NewGroupRepository()
    sessionRepo := repository.NewSessionRepository()
    msgRepo := repository.NewMessageRepository()
    fileRepo := repository.NewFileRepository()

    // 初始化 Service
    jwtSecret := app.Config("JWT_SECRET")
    jwtExpH := 72

    userSvc := service.NewUserService(userRepo, jwtSecret, jwtExpH)
    friendSvc := service.NewFriendService(friendRepo, userRepo)
    groupSvc := service.NewGroupService(groupRepo, sessionRepo, userRepo)
    msgSvc := service.NewMessageService(msgRepo, sessionRepo, hub)
    fileSvc := service.NewFileService(fileRepo)

    // 注册依赖（GoFr 依赖注入）
    app.AddDependency(userSvc, "userService")
    app.AddDependency(friendSvc, "friendService")
    app.AddDependency(groupSvc, "groupService")
    app.AddDependency(msgSvc, "messageService")
    app.AddDependency(fileSvc, "fileService")
    app.AddDependency(hub, "wsHub")

    // 注册中间件
    app.UseMiddleware(middleware.CORS)
    app.UseMiddleware(middleware.RequestLog)

    // 注册路由
    router.Register(app, hub)

    app.Run()
}
```

### 10.2 路由注册

`internal/router/router.go`：

```go
package router

import (
    "chat-app/internal/handler"
    "chat-app/internal/middleware"
    "chat-app/internal/ws"

    "gofr.dev/pkg/gofr"
)

func Register(app *gofr.App, hub *ws.Hub) {
    // 用户相关（无需鉴权）
    app.POST("/api/v1/auth/register", handler.Register)
    app.POST("/api/v1/auth/login", handler.Login)

    // 以下路由需要鉴权
    auth := middleware.Auth

    // 用户
    app.GET("/api/v1/users/profile", auth, handler.GetProfile)
    app.PUT("/api/v1/users/profile", auth, handler.UpdateProfile)
    app.GET("/api/v1/users/search", auth, handler.SearchUser)

    // 好友
    app.GET("/api/v1/friends", auth, handler.ListFriends)
    app.POST("/api/v1/friends/apply", auth, handler.ApplyFriend)
    app.POST("/api/v1/friends/approve", auth, handler.ApproveFriend)
    app.POST("/api/v1/friends/reject", auth, handler.RejectFriend)
    app.DELETE("/api/v1/friends/{friendID}", auth, handler.RemoveFriend)
    app.GET("/api/v1/friends/requests", auth, handler.ListPendingRequests)

    app.GET("/api/v1/friend-groups", auth, handler.ListFriendGroups)
    app.POST("/api/v1/friend-groups", auth, handler.CreateFriendGroup)
    app.DELETE("/api/v1/friend-groups/{groupID}", auth, handler.DeleteFriendGroup)

    // 群组
    app.POST("/api/v1/groups", auth, handler.CreateGroup)
    app.GET("/api/v1/groups/{groupID}", auth, handler.GetGroup)
    app.PUT("/api/v1/groups/{groupID}", auth, handler.UpdateGroup)
    app.DELETE("/api/v1/groups/{groupID}", auth, handler.DismissGroup)
    app.GET("/api/v1/groups/{groupID}/members", auth, handler.ListGroupMembers)
    app.POST("/api/v1/groups/{groupID}/members", auth, handler.AddGroupMember)
    app.DELETE("/api/v1/groups/{groupID}/members/{userID}", auth, handler.RemoveGroupMember)

    // 会话
    app.GET("/api/v1/sessions", auth, handler.ListSessions)
    app.GET("/api/v1/sessions/{sessionID}", auth, handler.GetSession)
    app.PUT("/api/v1/sessions/{sessionID}/mute", auth, handler.MuteSession)

    // 消息
    app.POST("/api/v1/messages", auth, handler.SendMessage)
    app.GET("/api/v1/messages/history", auth, handler.GetMessageHistory)
    app.POST("/api/v1/messages/read", auth, handler.MarkMessageRead)
    app.GET("/api/v1/messages/unread-count", auth, handler.GetUnreadCount)

    // 文件
    app.POST("/api/v1/files/upload", auth, handler.UploadFile)
    app.GET("/api/v1/files/{fileID}", auth, handler.GetFile)

    // WebSocket
    app.GET("/ws", auth, handler.WSHandler(hub))
}
```

### 10.3 Handler 示例

`internal/handler/user_handler.go`：

```go
package handler

import (
    "net/http"

    "chat-app/internal/pkg/response"
    "chat-app/internal/service"

    "gofr.dev/pkg/gofr"
)

func Register(c *gofr.Context) (interface{}, error) {
    var req service.RegisterRequest
    if err := c.Bind(&req); err != nil {
        return nil, response.BadRequest("invalid request body: " + err.Error())
    }

    svc := c.Get("userService").(service.UserService)
    user, token, err := svc.Register(c, &req)
    if err != nil {
        return nil, response.FromError(err)
    }

    return response.OK(map[string]interface{}{
        "user":  user,
        "token": token,
    }), nil
}

func Login(c *gofr.Context) (interface{}, error) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := c.Bind(&req); err != nil {
        return nil, response.BadRequest("invalid request body")
    }

    svc := c.Get("userService").(service.UserService)
    user, token, err := svc.Login(c, req.Username, req.Password)
    if err != nil {
        return nil, response.FromError(err)
    }

    return response.OK(map[string]interface{}{
        "user":  user,
        "token": token,
    }), nil
}

func GetProfile(c *gofr.Context) (interface{}, error) {
    userID := c.Get("userID").(int64)
    svc := c.Get("userService").(service.UserService)
    user, err := svc.GetProfile(c, userID)
    if err != nil {
        return nil, response.FromError(err)
    }
    return response.OK(user), nil
}

func UpdateProfile(c *gofr.Context) (interface{}, error) {
    userID := c.Get("userID").(int64)

    var req struct {
        Nickname  string `json:"nickname"`
        AvatarURL string `json:"avatar_url"`
        Gender    int16  `json:"gender"`
        Signature string `json:"signature"`
        Region    string `json:"region"`
    }
    if err := c.Bind(&req); err != nil {
        return nil, response.BadRequest("invalid request body")
    }

    user := &model.User{
        UserID:    userID,
        Nickname:  req.Nickname,
        AvatarURL: req.AvatarURL,
        Gender:    req.Gender,
        Signature: req.Signature,
        Region:    req.Region,
    }

    svc := c.Get("userService").(service.UserService)
    if err := svc.UpdateProfile(c, user); err != nil {
        return nil, response.FromError(err)
    }
    return response.OK(nil), nil
}
```

`internal/handler/message_handler.go`：

```go
package handler

import (
    "strconv"

    "chat-app/internal/pkg/response"
    "chat-app/internal/service"

    "gofr.dev/pkg/gofr"
)

func SendMessage(c *gofr.Context) (interface{}, error) {
    userID := c.Get("userID").(int64)

    var req service.SendMessageRequest
    if err := c.Bind(&req); err != nil {
        return nil, response.BadRequest("invalid request body")
    }

    svc := c.Get("messageService").(service.MessageService)
    msg, err := svc.Send(c, userID, &req)
    if err != nil {
        return nil, response.FromError(err)
    }
    return response.OK(msg), nil
}

func GetMessageHistory(c *gofr.Context) (interface{}, error) {
    userID := c.Get("userID").(int64)

    sessionID, err := strconv.ParseInt(c.PathParam("session_id"), 10, 64)
    if err != nil {
        return nil, response.BadRequest("invalid session_id")
    }

    beforeID, _ := strconv.ParseInt(c.Param("before_id"), 10, 64)
    limit, _ := strconv.Atoi(c.Param("limit"))
    if limit == 0 {
        limit = 50
    }

    svc := c.Get("messageService").(service.MessageService)
    messages, err := svc.History(c, userID, sessionID, beforeID, limit)
    if err != nil {
        return nil, response.FromError(err)
    }
    return response.OK(messages), nil
}

func GetUnreadCount(c *gofr.Context) (interface{}, error) {
    userID := c.Get("userID").(int64)
    svc := c.Get("messageService").(service.MessageService)
    counts, err := svc.UnreadCount(c, userID)
    if err != nil {
        return nil, response.FromError(err)
    }
    return response.OK(counts), nil
}
```

---

## 11. WebSocket 实时通信

### 11.1 Hub（连接管理中心）

`internal/ws/hub.go`：

```go
package ws

import (
    "encoding/json"
    "sync"
)

// WSMessage WebSocket 推送的消息结构
type WSMessage struct {
    Type string      `json:"type"` // new_message, friend_request, friend_online, etc.
    Data interface{} `json:"data"`
}

// Hub 管理所有 WebSocket 连接和会话路由
type Hub struct {
    // 用户ID -> 该用户的连接集合（一个用户可能多端登录）
    clients map[int64]map[*Client]bool

    // 会话ID -> 用户ID集合（用于群发）
    sessions map[int64]map[int64]bool

    // 注册/注销通道
    register   chan *Client
    unregister chan *Client

    // 广播通道
    broadcast chan *broadcastMsg

    mu sync.RWMutex
}

type broadcastMsg struct {
    sessionID int64
    message   *WSMessage
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[int64]map[*Client]bool),
        sessions:   make(map[int64]map[int64]bool),
        register:   make(chan *Client, 256),
        unregister: make(chan *Client, 256),
        broadcast:  make(chan *broadcastMsg, 256),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            if h.clients[client.userID] == nil {
                h.clients[client.userID] = make(map[*Client]bool)
            }
            h.clients[client.userID][client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if conns, ok := h.clients[client.userID]; ok {
                if _, exists := conns[client]; exists {
                    delete(conns, client)
                    close(client.send)
                    if len(conns) == 0 {
                        delete(h.clients, client.userID)
                    }
                }
            }
            h.mu.Unlock()

        case msg := <-h.broadcast:
            h.mu.RLock()
            userIDs := h.sessions[msg.sessionID]
            h.mu.RUnlock()

            payload, _ := json.Marshal(msg.message)
            for uid := range userIDs {
                h.sendToUser(uid, payload)
            }
        }
    }
}

// RegisterSession 注册用户到会话
func (h *Hub) RegisterSession(sessionID, userID int64) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if h.sessions[sessionID] == nil {
        h.sessions[sessionID] = make(map[int64]bool)
    }
    h.sessions[sessionID][userID] = true
}

// UnregisterSession 从会话注销用户
func (h *Hub) UnregisterSession(sessionID, userID int64) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if users, ok := h.sessions[sessionID]; ok {
        delete(users, userID)
        if len(users) == 0 {
            delete(h.sessions, sessionID)
        }
    }
}

// BroadcastToSession 向会话所有在线成员推送消息
func (h *Hub) BroadcastToSession(sessionID int64, msg *WSMessage) {
    h.broadcast <- &broadcastMsg{sessionID: sessionID, message: msg}
}

// SendToUser 向指定用户推送消息（所有端）
func (h *Hub) SendToUser(userID int64, msg *WSMessage) {
    payload, _ := json.Marshal(msg)
    h.sendToUser(userID, payload)
}

func (h *Hub) sendToUser(userID int64, payload []byte) {
    h.mu.RLock()
    conns := h.clients[userID]
    h.mu.RUnlock()

    for client := range conns {
        select {
        case client.send <- payload:
        default:
            // 发送缓冲已满，关闭连接
            h.unregister <- client
        }
    }
}

func (h *Hub) IsOnline(userID int64) bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    conns, ok := h.clients[userID]
    return ok && len(conns) > 0
}
```

### 11.2 Client（单连接封装）

`internal/ws/client.go`：

```go
package ws

import (
    "time"

    "github.com/gorilla/websocket"
)

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = 30 * time.Second
    maxMessageSize = 4096
)

type Client struct {
    hub     *Hub
    conn    *websocket.Conn
    send    chan []byte
    userID  int64
}

func NewClient(hub *Hub, conn *websocket.Conn, userID int64) *Client {
    return &Client{
        hub:    hub,
        conn:   conn,
        send:   make(chan []byte, 256),
        userID: userID,
    }
}

func (c *Client) ReadPump(handler MessageHandler) {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        handler.Handle(c, message)
    }
}

func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

### 11.3 WebSocket Handler

`internal/handler/ws_handler.go`：

```go
package handler

import (
    "net/http"
    "strconv"

    "chat-app/internal/ws"

    "github.com/gorilla/websocket"
    "gofr.dev/pkg/gofr"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  10240,
    WriteBufferSize: 10240,
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境需校验 Origin
    },
}

func WSHandler(hub *ws.Hub) func(c *gofr.Context) (interface{}, error) {
    return func(c *gofr.Context) (interface{}, error) {
        userID := c.Get("userID").(int64)

        // 升级 HTTP 连接为 WebSocket
        conn, err := upgrader.Upgrade(c.ResponseWriter, c.Request, nil)
        if err != nil {
            return nil, err
        }

        client := ws.NewClient(hub, conn, userID)
        hub.register <- client

        // 注册用户到其所有会话
        // TODO: 从 sessionRepo 加载该用户的所有 sessionID，调用 hub.RegisterSession

        go client.WritePump()
        go client.ReadPump(ws.NewDefaultMessageHandler(hub, c))

        return nil, nil
    }
}
```

---

## 12. 中间件

### 12.1 JWT 鉴权中间件

`internal/middleware/auth.go`：

```go
package middleware

import (
    "strings"

    "chat-app/internal/pkg/jwt"
    "chat-app/internal/pkg/response"

    "gofr.dev/pkg/gofr"
)

func Auth(c *gofr.Context) (interface{}, error) {
    authHeader := c.Request.Header.Get("Authorization")
    if authHeader == "" {
        return nil, response.Unauthorized("missing authorization header")
    }

    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 || parts[0] != "Bearer" {
        return nil, response.Unauthorized("invalid authorization format")
    }

    secret := c.Config("JWT_SECRET")
    claims, err := jwt.Parse(parts[1], secret)
    if err != nil {
        return nil, response.Unauthorized("invalid or expired token")
    }

    // 将用户信息注入 context
    c.Set("userID", claims.UserID)
    c.Set("username", claims.Username)

    return nil, nil // 返回 nil 表示继续执行后续 handler
}
```

> **注意**：GoFr 中间件返回 `nil, nil` 时表示继续执行下一个 handler/路由。如需中断，返回非 nil error。

### 12.2 CORS 中间件

`internal/middleware/cors.go`：

```go
package middleware

import (
    "gofr.dev/pkg/gofr"
)

func CORS(c *gofr.Context) (interface{}, error) {
    c.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
    c.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    c.ResponseWriter.Header().Set("Access-Control-Max-Age", "86400")

    if c.Request.Method == "OPTIONS" {
        c.ResponseWriter.WriteHeader(204)
        return nil, nil
    }
    return nil, nil
}
```

---

## 13. 错误处理与统一响应

### 13.1 错误码定义

`internal/pkg/errcode/errcode.go`：

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

### 13.2 统一响应封装

`internal/pkg/response/response.go`：

```go
package response

import (
    "errors"
    "net/http"

    "chat-app/internal/pkg/errcode"
)

type Response struct {
    Code    errcode.ErrCode `json:"code"`
    Message string          `json:"message"`
    Data    interface{}     `json:"data,omitempty"`
}

func OK(data interface{}) *Response {
    return &Response{
        Code:    errcode.Success,
        Message: "success",
        Data:    data,
    }
}

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

// APIError 实现 error 接口，并携带 HTTP 状态码
type APIError struct {
    Code       errcode.ErrCode
    Msg        string
    HTTPStatus int
}

func (e *APIError) Error() string { return e.Msg }

// FromError 将 service 层 error 转换为 APIError
func FromError(err error) error {
    if err == nil {
        return nil
    }

    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr
    }

    // 根据已知错误类型映射
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

---

## 14. API 接口清单

### 14.1 认证

| Method | Path | 说明 | 鉴权 |
|--------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | ❌ |
| POST | `/api/v1/auth/login` | 用户登录 | ❌ |

**注册请求**：
```json
{
  "username": "alice",
  "password": "secret123",
  "nickname": "Alice"
}
```

**注册响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user": {
      "user_id": 1,
      "username": "alice",
      "nickname": "Alice"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### 14.2 用户

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/users/profile` | 获取当前用户资料 |
| PUT | `/api/v1/users/profile` | 更新当前用户资料 |
| GET | `/api/v1/users/search?keyword=xxx` | 搜索用户 |

### 14.3 好友

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/friends` | 好友列表 |
| POST | `/api/v1/friends/apply` | 申请添加好友 |
| POST | `/api/v1/friends/approve` | 同意好友申请 |
| POST | `/api/v1/friends/reject` | 拒绝好友申请 |
| DELETE | `/api/v1/friends/{friendID}` | 删除好友 |
| GET | `/api/v1/friends/requests` | 待处理申请列表 |
| GET | `/api/v1/friend-groups` | 好友分组列表 |
| POST | `/api/v1/friend-groups` | 创建好友分组 |
| DELETE | `/api/v1/friend-groups/{groupID}` | 删除好友分组 |

**申请好友请求**：
```json
{
  "friend_id": 2,
  "remark": "老王",
  "group_id": 1,
  "message": "我是 Alice"
}
```

### 14.4 群组

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/v1/groups` | 创建群组 |
| GET | `/api/v1/groups/{groupID}` | 获取群信息 |
| PUT | `/api/v1/groups/{groupID}` | 更新群信息 |
| DELETE | `/api/v1/groups/{groupID}` | 解散群组 |
| GET | `/api/v1/groups/{groupID}/members` | 群成员列表 |
| POST | `/api/v1/groups/{groupID}/members` | 邀请入群 |
| DELETE | `/api/v1/groups/{groupID}/members/{userID}` | 移除群成员 |

### 14.5 会话

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/sessions` | 会话列表 |
| GET | `/api/v1/sessions/{sessionID}` | 会话详情 |
| PUT | `/api/v1/sessions/{sessionID}/mute` | 免打扰开关 |

### 14.6 消息

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/v1/messages` | 发送消息 |
| GET | `/api/v1/messages/history?session_id=1&before_id=100&limit=50` | 历史消息 |
| POST | `/api/v1/messages/read` | 标记已读 |
| GET | `/api/v1/messages/unread-count` | 未读消息数 |

**发送消息请求**：
```json
{
  "session_id": 1,
  "message_type": 1,
  "content": "你好！",
  "reply_to_msg_id": null
}
```

### 14.7 文件

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/v1/files/upload` | 上传文件（multipart） |
| GET | `/api/v1/files/{fileID}` | 获取文件信息 |

### 14.8 WebSocket

| Path | 说明 |
|------|------|
| `ws://host:8000/ws` | WebSocket 连接（Header: `Authorization: Bearer <token>`） |

**客户端→服务端消息**：
```json
{
  "type": "ping",
  "data": {}
}
```

**服务端→客户端消息**：
```json
{
  "type": "new_message",
  "data": {
    "message_id": 100,
    "session_id": 1,
    "sender_id": 2,
    "message_type": 1,
    "content": "你好！",
    "sent_at": "2026-07-04T10:30:00Z"
  }
}
```

**消息类型枚举**：
- `new_message` - 新消息
- `friend_request` - 好友申请
- `friend_approved` - 好友申请通过
- `user_online` - 用户上线
- `user_offline` - 用户下线
- `message_read` - 消息已读回执
- `group_member_joined` - 群成员加入
- `group_member_left` - 群成员退出

---

## 15. 测试

### 15.1 单元测试示例

`internal/service/user_service_test.go`：

```go
package service

import (
    "context"
    "testing"

    "chat-app/internal/model"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock UserRepository
type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    args := m.Called(ctx, username)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

// ... 其他方法 mock

func TestRegister_Success(t *testing.T) {
    repo := new(MockUserRepo)
    svc := NewUserService(repo, "test-secret", 72)

    repo.On("GetByUsername", mock.Anything, "alice").Return(nil, nil)
    repo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

    user, token, err := svc.Register(context.Background(), &RegisterRequest{
        Username: "alice",
        Password: "secret123",
        Nickname: "Alice",
    })

    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.NotEmpty(t, token)
    repo.AssertExpectations(t)
}

func TestRegister_UsernameExists(t *testing.T) {
    repo := new(MockUserRepo)
    svc := NewUserService(repo, "test-secret", 72)

    repo.On("GetByUsername", mock.Anything, "alice").Return(&model.User{UserID: 1}, nil)

    _, _, err := svc.Register(context.Background(), &RegisterRequest{
        Username: "alice",
        Password: "secret123",
    })

    assert.Equal(t, ErrUsernameExists, err)
}
```

### 15.2 集成测试

使用 `testcontainers-go` 启动真实 PostgreSQL：

```go
package repository_test

import (
    "context"
    "testing"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupTestDB(t *testing.T) *sql.DB {
    ctx := context.Background()
    pgContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.WithInitScripts("../../migrations/001_init_schema.up.sql"),
    )
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() {
        pgContainer.Terminate(ctx)
    })

    connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
    db, _ := sql.Open("postgres", connStr)
    return db
}
```

---

## 16. 部署

### 16.1 Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /chat-app ./cmd/server

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /chat-app .
COPY migrations ./migrations

EXPOSE 8000
CMD ["./chat-app"]
```

### 16.2 Makefile

```makefile
.PHONY: build run test lint migrate-up migrate-down docker-build docker-up

build:
	go build -o bin/chat-app ./cmd/server

run:
	go run cmd/server/main.go

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run

migrate-up:
	migrate -path migrations -database "postgres://chat:chat123@localhost:5432/chatapp?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://chat:chat123@localhost:5432/chatapp?sslmode=disable" down 1

docker-build:
	docker build -t chat-app:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin/
```

---

## 17. 开发约定与注意事项

### 17.1 代码规范

| 规则 | 说明 |
|------|------|
| 包命名 | 全小写，单数（`user` 而非 `users`） |
| 文件命名 | 全小写+下划线（`user_repo.go`） |
| 接口命名 | 以 `er` 结尾（`UserRepository`）或描述行为（`UserService`） |
| 错误变量 | 以 `Err` 前缀（`ErrNotFound`） |
| 常量命名 | 驼峰式（`MessageTypeText`） |
| 注释 | 所有导出函数必须有 GoDoc 注释 |

### 17.2 数据库注意事项

1. **时间字段统一使用 `TIMESTAMPTZ`**，避免时区问题
2. **主键统一使用 `BIGSERIAL`**，预留扩展空间
3. **软删除 vs 硬删除**：本设计采用硬删除（`ON DELETE CASCADE`），如需保留历史数据可加 `deleted_at` 字段
4. **大表索引**：`messages` 表会快速增长，索引 `(session_id, sent_at DESC)` 是关键
5. **事务边界**：消息发送、好友申请通过等操作必须在事务中完成
6. **N+1 查询**：列表接口避免循环查询，使用 `IN` / `JOIN` 批量加载

### 17.3 WebSocket 注意事项

1. **心跳机制**：客户端每 30s 发送 ping，服务端 60s 无响应则断开
2. **多端登录**：同一用户允许多个 WebSocket 连接（手机+电脑）
3. **消息可靠性**：WebSocket 仅用于实时推送，消息持久化以 HTTP 接口为准。客户端收到推送后应通过 HTTP 拉取完整消息
4. **连接恢复**：客户端断线重连后，应调用 `/messages/history?before_id=<本地最大ID>` 拉取离线消息

### 17.4 安全注意事项

1. **密码存储**：必须使用 bcrypt 哈希，cost ≥ 10
2. **SQL 注入**：所有查询使用参数化（`$1, $2`），禁止字符串拼接
3. **JWT 密钥**：生产环境必须使用高强度随机密钥（≥32 字节）
4. **文件上传**：校验 MIME 类型、限制大小、重命名存储（避免路径穿越）
5. **限流**：登录接口、注册接口建议加 IP 限流（Redis 实现）

### 17.5 性能优化建议

| 场景 | 优化方案 |
|------|----------|
| 消息表过大 | 按 `sent_at` 月份分区 |
| 会话列表查询慢 | 在 `session_participants` 上加 `(user_id, joined_at DESC)` 索引 |
| 未读计数实时性 | 用 Redis 维护未读计数，避免每次查询数据库 |
| 群消息广播 | 大群（>500人）采用异步队列批量推送 |
| 历史消息分页 | 使用 `WHERE message_id < $before_id` 游标分页，避免 OFFSET |

### 17.6 开发流程

```
1. 创建分支：feature/xxx 或 fix/xxx
2. 编写代码 + 单元测试
3. 本地运行：make test && make lint
4. 提交 PR，CI 通过后合并
5. 合并到 main 后自动部署到 staging
6. 手动验证后部署到 production
```

---

## 附录 A：GoFr 框架快速参考

### A.1 创建应用

```go
app := gofr.New()
app.Run()
```

### A.2 注册路由

```go
app.GET("/path", handler)
app.POST("/path", handler)
app.PUT("/path/{id}", handler)
app.DELETE("/path/{id}", handler)
```

### A.3 读取配置

```go
dbHost := c.Config("DB_HOST")
port := c.Config("HTTP_PORT")
```

### A.4 数据库访问

```go
db := c.DB  // *sql.DB
rows, err := db.QueryContext(ctx, "SELECT ...")
```

### A.5 Redis 访问

```go
rdb := c.Redis  // *redis.Client
rdb.Set(ctx, "key", "value", 0)
```

### A.6 日志

```go
c.Logger.Infof("user %d logged in", userID)
c.Logger.Errorf("db error: %v", err)
```

### A.7 依赖注入

```go
// 注册
app.AddDependency(myService, "myService")

// 获取
svc := c.Get("myService").(MyService)
```

### A.8 请求绑定

```go
var req MyRequest
if err := c.Bind(&req); err != nil {
    return nil, err
}
```

---

## 附录 B：常用 psql 命令

```bash
# 连接数据库
psql -h localhost -U chat -d chatapp

# 查看所有表
\dt

# 查看表结构
\d users

# 查看索引
\di

# 执行 SQL 文件
\i migrations/001_init_schema.up.sql

# 导出数据
pg_dump -h localhost -U chat chatapp > backup.sql

# 查看连接数
SELECT count(*) FROM pg_stat_activity WHERE datname = 'chatapp';
```

---

**文档结束**
