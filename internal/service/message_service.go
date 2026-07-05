package service

import (
	"context"
	"errors"

	"chatgoo/internal/model"
	"chatgoo/internal/repository"
	"chatgoo/internal/ws"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrNotParticipant  = errors.New("user is not a participant of this session")
	ErrEmptyMessage    = errors.New("message content cannot be empty")
)

// SendMessageRequest holds the message sending payload.
type SendMessageRequest struct {
	SessionID    int64  `json:"session_id"`
	MessageType  int16  `json:"message_type"`
	Content      string `json:"content"`
	FileID       *int64 `json:"file_id"`
	ReplyToMsgID *int64 `json:"reply_to_msg_id"`
}

// MessageService defines message business operations.
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

// NewMessageService creates a MessageService.
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
	session, err := s.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	if req.MessageType == model.MessageTypeText && req.Content == "" {
		return nil, ErrEmptyMessage
	}

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

	s.hub.BroadcastToSession(req.SessionID, &ws.WSMessage{
		Type: "new_message",
		Data: msg,
	})

	return msg, nil
}

func (s *messageService) History(ctx context.Context, userID, sessionID, beforeID int64, limit int) ([]*model.Message, error) {
	return s.msgRepo.ListBySession(ctx, sessionID, beforeID, limit)
}

func (s *messageService) MarkRead(ctx context.Context, userID, sessionID, msgID int64) error {
	return s.sessionRepo.UpdateLastRead(ctx, sessionID, userID, msgID)
}

func (s *messageService) UnreadCount(ctx context.Context, userID int64) (map[int64]int, error) {
	return s.msgRepo.GetUnreadCount(ctx, userID)
}
