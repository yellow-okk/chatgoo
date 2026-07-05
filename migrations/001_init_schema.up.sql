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
-- 8. 文件表（必须在 messages 之前创建）
-- ============================================================
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
-- 9. 消息表
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
