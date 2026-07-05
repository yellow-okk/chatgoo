package repository

import (
	"context"

	"chatgoo/internal/model"

	"gofr.dev/pkg/gofr/container"
)

// MessageRepository defines data access methods for messages.
type MessageRepository interface {
	Create(ctx context.Context, msg *model.Message) error
	GetByID(ctx context.Context, msgID int64) (*model.Message, error)
	ListBySession(ctx context.Context, sessionID int64, beforeID int64, limit int) ([]*model.Message, error)
	UpdateReadStatus(ctx context.Context, msgID, userID int64) error
	GetUnreadCount(ctx context.Context, userID int64) (map[int64]int, error)
}

type messageRepository struct {
	db container.DB
}

// NewMessageRepository creates a MessageRepository.
func NewMessageRepository(db container.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, msg *model.Message) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

	_, err = tx.ExecContext(ctx, `
		UPDATE chat_sessions
		SET last_message_id = $1, last_message_at = $2
		WHERE session_id = $3
	`, msg.MessageID, msg.SentAt, msg.SessionID)
	if err != nil {
		return err
	}

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

func (r *messageRepository) GetByID(ctx context.Context, msgID int64) (*model.Message, error) {
	var m model.Message
	query := `
		SELECT message_id, session_id, sender_id, message_type, content,
		       file_id, reply_to_msg_id, sent_at
		FROM messages WHERE message_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, msgID).Scan(
		&m.MessageID, &m.SessionID, &m.SenderID, &m.MessageType,
		&m.Content, &m.FileID, &m.ReplyToMsgID, &m.SentAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepository) ListBySession(ctx context.Context, sessionID int64, beforeID int64, limit int) ([]*model.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var query string
	var args []any
	if beforeID > 0 {
		query = `
			SELECT message_id, session_id, sender_id, message_type, content,
			       file_id, reply_to_msg_id, sent_at
			FROM messages
			WHERE session_id = $1 AND message_id < $2
			ORDER BY message_id DESC
			LIMIT $3
		`
		args = []any{sessionID, beforeID, limit}
	} else {
		query = `
			SELECT message_id, session_id, sender_id, message_type, content,
			       file_id, reply_to_msg_id, sent_at
			FROM messages
			WHERE session_id = $1
			ORDER BY message_id DESC
			LIMIT $2
		`
		args = []any{sessionID, limit}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.MessageID, &m.SessionID, &m.SenderID, &m.MessageType,
			&m.Content, &m.FileID, &m.ReplyToMsgID, &m.SentAt); err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}
	return messages, nil
}

func (r *messageRepository) UpdateReadStatus(ctx context.Context, msgID, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO message_status (message_id, user_id, read_status, read_at)
		VALUES ($1, $2, 1, CURRENT_TIMESTAMP)
		ON CONFLICT (message_id, user_id)
		DO UPDATE SET read_status = 1, read_at = CURRENT_TIMESTAMP
	`, msgID, userID)
	return err
}

func (r *messageRepository) GetUnreadCount(ctx context.Context, userID int64) (map[int64]int, error) {
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
	rows, err := r.db.QueryContext(ctx, query, userID)
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
