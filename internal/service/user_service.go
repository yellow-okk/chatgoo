package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"chatgoo/internal/model"
	"chatgoo/internal/pkg/hash"
	"chatgoo/internal/pkg/jwt"
	"chatgoo/internal/repository"
)

var (
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTooShort   = errors.New("username must be at least 3 characters")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
)

// RegisterRequest holds the registration payload.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// UserService defines user business operations.
type UserService interface {
	Register(ctx context.Context, req *RegisterRequest) (*model.User, string, error)
	Login(ctx context.Context, username, password string) (*model.User, string, error)
	GetProfile(ctx context.Context, userID int64) (*model.User, error)
	UpdateProfile(ctx context.Context, user *model.User) error
	SearchUser(ctx context.Context, keyword string) ([]*model.User, error)
}

type userService struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtExpH   int
}

// NewUserService creates a new UserService.
func NewUserService(userRepo repository.UserRepository, jwtSecret string, jwtExpH int) UserService {
	return &userService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpH:   jwtExpH,
	}
}

func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*model.User, string, error) {
	if len(strings.TrimSpace(req.Username)) < 3 {
		return nil, "", ErrUsernameTooShort
	}
	if len(req.Password) < 6 {
		return nil, "", ErrPasswordTooShort
	}

	existing, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		return nil, "", ErrUsernameExists
	}

	hashed, err := hash.Bcrypt(req.Password)
	if err != nil {
		return nil, "", err
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: hashed,
		Nickname:     req.Nickname,
		Status:       model.UserStatusOffline,
	}
	if user.Nickname == "" {
		user.Nickname = req.Username
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := jwt.Generate(user.UserID, user.Username, s.jwtSecret, time.Duration(s.jwtExpH)*time.Hour)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *userService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return nil, "", ErrInvalidCredentials
	}

	if !hash.VerifyBcrypt(password, user.PasswordHash) {
		return nil, "", ErrInvalidCredentials
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID); err != nil {
		return nil, "", err
	}

	token, err := jwt.Generate(user.UserID, user.Username, s.jwtSecret, time.Duration(s.jwtExpH)*time.Hour)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *userService) UpdateProfile(ctx context.Context, user *model.User) error {
	return s.userRepo.Update(ctx, user)
}

func (s *userService) SearchUser(ctx context.Context, keyword string) ([]*model.User, error) {
	return nil, nil
}
