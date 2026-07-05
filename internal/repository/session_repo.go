package repository

import (
	"context"
	"database/sql"

	"chatgoo/internal/model"

	"gofr.dev/pkg/gofr/container"
)

// SessionRepository defines data access methods for chat sessions.
type SessionRepository interface {
	GetOrCreatePrivateSession(ctx context.Context, userA, userB int64) (int64, error)
	GetOrCreateGroupSession(ctx context.Context, groupID int64) (int64, error)
	GetByID(ctx context.Context, sessionID int64) (*model.ChatSession, error)
	ListByUser(ctx context.Context, userID int64) ([]*model.ChatSession, error)
	UpdateLastRead(ctx context.Context, sessionID, userID, msgID int64) error
	Mute(ctx context.Context, sessionID, userID int64, muted bool) error
}

type sessionRepository struct {
	db container.DB
}

// NewSessionRepository creates a SessionRepository.
func NewSessionRepository(db container.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) GetOrCreatePrivateSession(ctx context.Context, userA, userB int64) (int64, error) {
	// Try to find existing private session between the two users
	query := `
		SELECT cs.session_id FROM chat_sessions cs
		WHERE cs.session_type = $1
		  AND cs.target_user_id = $2
		  AND EXISTS (
		      SELECT 1 FROM session_participants sp
		      WHERE sp.session_id = cs.session_id AND sp.user_id = $3
		  )
		LIMIT 1
	`
	var sessionID int64
	err := r.db.QueryRowContext(ctx, query, model.SessionTypePrivate, userB, userA).Scan(&sessionID)
	if err == nil {
		return sessionID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Also check the reverse direction
	err = r.db.QueryRowContext(ctx, query, model.SessionTypePrivate, userA, userB).Scan(&sessionID)
	if err == nil {
		return sessionID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Create new session
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO chat_sessions (session_type, target_user_id) VALUES ($1, $2)
		RETURNING session_id
	`, model.SessionTypePrivate, userB).Scan(&sessionID)
	if err != nil {
		return 0, err
	}

	for _, uid := range []int64{userA, userB} {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO session_participants (session_id, user_id) VALUES ($1, $2)`, sessionID, uid)
		if err != nil {
			return 0, err
		}
	}

	return sessionID, tx.Commit()
}

func (r *sessionRepository) GetOrCreateGroupSession(ctx context.Context, groupID int64) (int64, error) {
	var sessionID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT session_id FROM chat_sessions WHERE session_type = $1 AND group_id = $2`,
		model.SessionTypeGroup, groupID).Scan(&sessionID)
	if err == nil {
		return sessionID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO chat_sessions (session_type, group_id) VALUES ($1, $2) RETURNING session_id
	`, model.SessionTypeGroup, groupID).Scan(&sessionID)
	if err != nil {
		return 0, err
	}

	// Add all group members as session participants
	_, err = tx.ExecContext(ctx, `
		INSERT INTO session_participants (session_id, user_id)
		SELECT $1, user_id FROM group_members WHERE group_id = $2
	`, sessionID, groupID)
	if err != nil {
		return 0, err
	}

	return sessionID, tx.Commit()
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID int64) (*model.ChatSession, error) {
	var s model.ChatSession
	query := `
		SELECT session_id, session_type, target_user_id, group_id,
		       last_message_id, last_message_at, created_at, updated_at
		FROM chat_sessions WHERE session_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&s.SessionID, &s.SessionType, &s.TargetUserID, &s.GroupID,
		&s.LastMessageID, &s.LastMessageAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *sessionRepository) ListByUser(ctx context.Context, userID int64) ([]*model.ChatSession, error) {
	query := `
		SELECT cs.session_id, cs.session_type, cs.target_user_id, cs.group_id,
		       cs.last_message_id, cs.last_message_at, cs.created_at, cs.updated_at
		FROM chat_sessions cs
		JOIN session_participants sp ON sp.session_id = cs.session_id
		WHERE sp.user_id = $1
		ORDER BY COALESCE(cs.last_message_at, cs.created_at) DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.ChatSession
	for rows.Next() {
		var s model.ChatSession
		if err := rows.Scan(&s.SessionID, &s.SessionType, &s.TargetUserID, &s.GroupID,
			&s.LastMessageID, &s.LastMessageAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

func (r *sessionRepository) UpdateLastRead(ctx context.Context, sessionID, userID, msgID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE session_participants SET last_read_msg_id = $1
		WHERE session_id = $2 AND user_id = $3
	`, msgID, sessionID, userID)
	return err
}

func (r *sessionRepository) Mute(ctx context.Context, sessionID, userID int64, muted bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE session_participants SET muted = $1
		WHERE session_id = $2 AND user_id = $3
	`, muted, sessionID, userID)
	return err
}
