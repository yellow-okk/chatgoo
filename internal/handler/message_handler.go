package handler

import (
	"strconv"

	"chatgoo/internal/middleware"
	"chatgoo/internal/pkg/response"
	"chatgoo/internal/service"

	"gofr.dev/pkg/gofr"
)

// SendMessage sends a message to a session.
func SendMessage(msgSvc service.MessageService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

		var req service.SendMessageRequest
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		msg, err := msgSvc.Send(c, userID, &req)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(msg), nil
	}
}

// GetMessageHistory returns message history for a session.
func GetMessageHistory(msgSvc service.MessageService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

		sessionID, err := strconv.ParseInt(c.Param("session_id"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid session_id")
		}

		beforeID, _ := strconv.ParseInt(c.Param("before_id"), 10, 64)
		limit, _ := strconv.Atoi(c.Param("limit"))
		if limit == 0 {
			limit = 50
		}

		messages, err := msgSvc.History(c, userID, sessionID, beforeID, limit)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(messages), nil
	}
}

// MarkMessageRead marks messages as read.
func MarkMessageRead(msgSvc service.MessageService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

		var req struct {
			SessionID int64 `json:"session_id"`
			MessageID int64 `json:"message_id"`
		}
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		if err := msgSvc.MarkRead(c, userID, req.SessionID, req.MessageID); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// GetUnreadCount returns unread message counts per session.
func GetUnreadCount(msgSvc service.MessageService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		counts, err := msgSvc.UnreadCount(c, userID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(counts), nil
	}
}
