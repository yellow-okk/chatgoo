package handler

import (
	"chatgoo/internal/middleware"
	"chatgoo/internal/model"
	"chatgoo/internal/pkg/response"
	"chatgoo/internal/service"

	"gofr.dev/pkg/gofr"
)

// Register handles user registration.
func Register(userSvc service.UserService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		var req service.RegisterRequest
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body: " + err.Error())
		}

		user, token, err := userSvc.Register(c, &req)
		if err != nil {
			return nil, response.FromError(err)
		}

		return response.OK(map[string]any{
			"user":  user,
			"token": token,
		}), nil
	}
}

// Login handles user authentication.
func Login(userSvc service.UserService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		user, token, err := userSvc.Login(c, req.Username, req.Password)
		if err != nil {
			return nil, response.FromError(err)
		}

		return response.OK(map[string]any{
			"user":  user,
			"token": token,
		}), nil
	}
}

// GetProfile returns the authenticated user's profile.
func GetProfile(userSvc service.UserService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		user, err := userSvc.GetProfile(c, userID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(user), nil
	}
}

// UpdateProfile updates the authenticated user's profile.
func UpdateProfile(userSvc service.UserService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

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

		if err := userSvc.UpdateProfile(c, user); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// SearchUser searches for users by keyword.
func SearchUser(userSvc service.UserService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		keyword := c.Param("keyword")
		users, err := userSvc.SearchUser(c, keyword)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(users), nil
	}
}
