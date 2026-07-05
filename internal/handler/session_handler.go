package handler

import (
	"strconv"

	"chatgoo/internal/middleware"
	"chatgoo/internal/pkg/response"
	"chatgoo/internal/repository"

	"gofr.dev/pkg/gofr"
)

// ListSessions returns the authenticated user's session list.
func ListSessions(sessionRepo repository.SessionRepository) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		sessions, err := sessionRepo.ListByUser(c, userID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(sessions), nil
	}
}

// GetSession returns session details.
func GetSession(sessionRepo repository.SessionRepository) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		sessionID, err := strconv.ParseInt(c.PathParam("sessionID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid sessionID")
		}

		session, err := sessionRepo.GetByID(c, sessionID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(session), nil
	}
}

// MuteSession toggles mute for a session.
func MuteSession(sessionRepo repository.SessionRepository) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		sessionID, err := strconv.ParseInt(c.PathParam("sessionID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid sessionID")
		}

		var req struct {
			Muted bool `json:"muted"`
		}
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		if err := sessionRepo.Mute(c, sessionID, userID, req.Muted); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}
