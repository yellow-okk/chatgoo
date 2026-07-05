package router

import (
	"chatgoo/internal/handler"
	"chatgoo/internal/repository"
	"chatgoo/internal/service"
	"chatgoo/internal/ws"

	"gofr.dev/pkg/gofr"
)

// Register sets up all HTTP routes.
func Register(
	app *gofr.App,
	userSvc service.UserService,
	friendSvc service.FriendService,
	groupSvc service.GroupService,
	msgSvc service.MessageService,
	fileSvc service.FileService,
	sessionRepo repository.SessionRepository,
	hub *ws.Hub,
	jwtSecret string,
) {
	// Auth (public)
	app.POST("/api/v1/auth/register", handler.Register(userSvc))
	app.POST("/api/v1/auth/login", handler.Login(userSvc))

	// User
	app.GET("/api/v1/users/profile", handler.GetProfile(userSvc))
	app.PUT("/api/v1/users/profile", handler.UpdateProfile(userSvc))
	app.GET("/api/v1/users/search", handler.SearchUser(userSvc))

	// Friends
	app.GET("/api/v1/friends", handler.ListFriends(friendSvc))
	app.POST("/api/v1/friends/apply", handler.ApplyFriend(friendSvc))
	app.POST("/api/v1/friends/approve", handler.ApproveFriend(friendSvc))
	app.POST("/api/v1/friends/reject", handler.RejectFriend(friendSvc))
	app.DELETE("/api/v1/friends/{friendID}", handler.RemoveFriend(friendSvc))
	app.GET("/api/v1/friends/requests", handler.ListPendingRequests(friendSvc))

	// Friend Groups
	app.GET("/api/v1/friend-groups", handler.ListFriendGroups(friendSvc))
	app.POST("/api/v1/friend-groups", handler.CreateFriendGroup(friendSvc))
	app.DELETE("/api/v1/friend-groups/{groupID}", handler.DeleteFriendGroup(friendSvc))

	// Groups
	app.POST("/api/v1/groups", handler.CreateGroup(groupSvc))
	app.GET("/api/v1/groups/{groupID}", handler.GetGroup(groupSvc))
	app.PUT("/api/v1/groups/{groupID}", handler.UpdateGroup(groupSvc))
	app.DELETE("/api/v1/groups/{groupID}", handler.DismissGroup(groupSvc))
	app.GET("/api/v1/groups/{groupID}/members", handler.ListGroupMembers(groupSvc))
	app.POST("/api/v1/groups/{groupID}/members", handler.AddGroupMember(groupSvc))
	app.DELETE("/api/v1/groups/{groupID}/members/{userID}", handler.RemoveGroupMember(groupSvc))

	// Sessions
	app.GET("/api/v1/sessions", handler.ListSessions(sessionRepo))
	app.GET("/api/v1/sessions/{sessionID}", handler.GetSession(sessionRepo))
	app.PUT("/api/v1/sessions/{sessionID}/mute", handler.MuteSession(sessionRepo))

	// Messages
	app.POST("/api/v1/messages", handler.SendMessage(msgSvc))
	app.GET("/api/v1/messages/history", handler.GetMessageHistory(msgSvc))
	app.POST("/api/v1/messages/read", handler.MarkMessageRead(msgSvc))
	app.GET("/api/v1/messages/unread-count", handler.GetUnreadCount(msgSvc))

	// Files
	app.GET("/api/v1/files/{fileID}", handler.GetFile(fileSvc))

	// WebSocket (auth via ?token= query param)
	app.WebSocket("/ws", handler.WSHandler(hub, jwtSecret))
}
