package main

import (
	"chatgoo/internal/middleware"
	"chatgoo/internal/repository"
	"chatgoo/internal/router"
	"chatgoo/internal/service"
	"chatgoo/internal/ws"

	"gofr.dev/pkg/gofr"
)

func main() {
	app := gofr.New()

	// WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// Database (GoFr auto-connects via DB_* env vars)
	db := app.GetSQL()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	friendRepo := repository.NewFriendRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	msgRepo := repository.NewMessageRepository(db)
	fileRepo := repository.NewFileRepository(db)

	// Services
	jwtSecret := app.Config.Get("JWT_SECRET")
	jwtExpH := 72

	userSvc := service.NewUserService(userRepo, jwtSecret, jwtExpH)
	friendSvc := service.NewFriendService(friendRepo, userRepo)
	groupSvc := service.NewGroupService(groupRepo, sessionRepo)
	msgSvc := service.NewMessageService(msgRepo, sessionRepo, hub)
	fileSvc := service.NewFileService(fileRepo)

	// Middleware (auth applied globally, public paths skipped)
	app.UseMiddleware(middleware.Auth(jwtSecret))

	// Routes
	router.Register(app, userSvc, friendSvc, groupSvc, msgSvc, fileSvc, sessionRepo, hub, jwtSecret)

	app.Run()
}
