package handler

import (
	"strconv"

	"chatgoo/internal/middleware"
	"chatgoo/internal/pkg/response"
	"chatgoo/internal/service"

	"gofr.dev/pkg/gofr"
)

// ListFriends returns the authenticated user's friend list.
func ListFriends(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		friends, err := friendSvc.ListFriends(c, userID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(friends), nil
	}
}

// ApplyFriend sends a friend request.
func ApplyFriend(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

		var req service.ApplyFriendRequest
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		if err := friendSvc.Apply(c, userID, &req); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// ApproveFriend approves a friend request.
func ApproveFriend(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

		var req struct {
			RelationID int64 `json:"relation_id"`
		}
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		if err := friendSvc.Approve(c, userID, req.RelationID); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// RejectFriend rejects a friend request.
func RejectFriend(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

		var req struct {
			RelationID int64 `json:"relation_id"`
		}
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		if err := friendSvc.Reject(c, userID, req.RelationID); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// RemoveFriend removes a friend.
func RemoveFriend(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		friendID, err := strconv.ParseInt(c.PathParam("friendID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid friendID")
		}

		if err := friendSvc.Remove(c, userID, friendID); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}

// ListPendingRequests returns pending friend requests.
func ListPendingRequests(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		requests, err := friendSvc.ListPendingRequests(c, userID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(requests), nil
	}
}

// ListFriendGroups returns the user's friend groups.
func ListFriendGroups(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		groups, err := friendSvc.ListGroups(c, userID)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(groups), nil
	}
}

// CreateFriendGroup creates a new friend group.
func CreateFriendGroup(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)

		var req struct {
			GroupName string `json:"group_name"`
		}
		if err := c.Bind(&req); err != nil {
			return nil, response.BadRequest("invalid request body")
		}

		group, err := friendSvc.CreateGroup(c, userID, req.GroupName)
		if err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(group), nil
	}
}

// DeleteFriendGroup deletes a friend group.
func DeleteFriendGroup(friendSvc service.FriendService) func(c *gofr.Context) (any, error) {
	return func(c *gofr.Context) (any, error) {
		userID := c.Request.Context().Value(middleware.UserIDKey).(int64)
		groupID, err := strconv.ParseInt(c.PathParam("groupID"), 10, 64)
		if err != nil {
			return nil, response.BadRequest("invalid groupID")
		}

		if err := friendSvc.DeleteGroup(c, userID, groupID); err != nil {
			return nil, response.FromError(err)
		}
		return response.OK(nil), nil
	}
}
