package handler

import (
	"strconv"

	"chatgoo/internal/middleware"
	"chatgoo/internal/model"
	"chatgoo/internal/pkg/response"
	"chatgoo/internal/service"

	"gofr.dev/pkg/gofr"
)

// CreateGroup creates a new chat group.
func CreateGroup(groupSvc service.GroupService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

		var req service.CreateGroupRequest
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		group, err := groupSvc.Create(c, userID, &req)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(group), nil
	}
}

// GetGroup returns group info.
func GetGroup(groupSvc service.GroupService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		groupID, err := strconv.ParseInt(c.PathParam("groupID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid groupID")
		}

		group, err := groupSvc.GetByID(c, groupID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(group), nil
	}
}

// UpdateGroup updates group info.
func UpdateGroup(groupSvc service.GroupService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		groupID, err := strconv.ParseInt(c.PathParam("groupID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid groupID")
		}

		var req struct {
			GroupName    string `json:"group_name"`
			AvatarURL    string `json:"avatar_url"`
			Announcement string `json:"announcement"`
			MaxMembers   int    `json:"max_members"`
		}
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		group := &model.GroupInfo{
			GroupID:      groupID,
			GroupName:    req.GroupName,
			AvatarURL:    req.AvatarURL,
			Announcement: req.Announcement,
			MaxMembers:   req.MaxMembers,
		}

		if err := groupSvc.Update(c, userID, group); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// DismissGroup disbands a group.
func DismissGroup(groupSvc service.GroupService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		groupID, err := strconv.ParseInt(c.PathParam("groupID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid groupID")
		}

		if err := groupSvc.Dismiss(c, userID, groupID); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// ListGroupMembers returns group member list.
func ListGroupMembers(groupSvc service.GroupService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		groupID, err := strconv.ParseInt(c.PathParam("groupID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid groupID")
		}

		members, err := groupSvc.ListMembers(c, groupID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(members), nil
	}
}

// AddGroupMember adds a user to a group.
func AddGroupMember(groupSvc service.GroupService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		groupID, err := strconv.ParseInt(c.PathParam("groupID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid groupID")
		}

		var req struct {
			UserID int64 `json:"user_id"`
		}
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		if err := groupSvc.AddMember(c, groupID, req.UserID); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// RemoveGroupMember removes a user from a group.
func RemoveGroupMember(groupSvc service.GroupService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		groupID, err := strconv.ParseInt(c.PathParam("groupID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid groupID")
		}
		userID, err := strconv.ParseInt(c.PathParam("userID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid userID")
		}

		if err := groupSvc.RemoveMember(c, groupID, userID); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}
