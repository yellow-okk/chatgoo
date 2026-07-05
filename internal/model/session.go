package model

import "time"

// ChatSession represents a conversation (private or group).
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

// SessionParticipant tracks a user's participation in a session.
type SessionParticipant struct {
	ID            int64     `json:"id" db:"id"`
	SessionID     int64     `json:"session_id" db:"session_id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	LastReadMsgID int64     `json:"last_read_msg_id" db:"last_read_msg_id"`
	Muted         bool      `json:"muted" db:"muted"`
	JoinedAt      time.Time `json:"joined_at" db:"joined_at"`
}

const (
	SessionTypePrivate int16 = 1
	SessionTypeGroup   int16 = 2
)
