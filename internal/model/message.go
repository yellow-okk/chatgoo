package model

import "time"

// Message represents a chat message.
type Message struct {
	MessageID    int64     `json:"message_id" db:"message_id"`
	SessionID    int64     `json:"session_id" db:"session_id"`
	SenderID     int64     `json:"sender_id" db:"sender_id"`
	MessageType  int16     `json:"message_type" db:"message_type"`
	Content      string    `json:"content" db:"content"`
	FileID       *int64    `json:"file_id" db:"file_id"`
	ReplyToMsgID *int64    `json:"reply_to_msg_id" db:"reply_to_msg_id"`
	SentAt       time.Time `json:"sent_at" db:"sent_at"`
}

// MessageStatus tracks read receipts.
type MessageStatus struct {
	ID         int64      `json:"id" db:"id"`
	MessageID  int64      `json:"message_id" db:"message_id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	ReadStatus int16      `json:"read_status" db:"read_status"`
	ReadAt     *time.Time `json:"read_at" db:"read_at"`
}

const (
	MessageTypeText  int16 = 1
	MessageTypeImage int16 = 2
	MessageTypeFile  int16 = 3
	MessageTypeVoice int16 = 4
)

const (
	ReadStatusUnread int16 = 0
	ReadStatusRead   int16 = 1
)
