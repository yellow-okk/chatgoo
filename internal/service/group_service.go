package service

import (
	"context"
	"errors"

	"chatgoo/internal/model"
	"chatgoo/internal/repository"
)

var (
	ErrGroupNotFound  = errors.New("group not found")
	ErrNotGroupOwner  = errors.New("only group owner can perform this action")
	ErrAlreadyMember  = errors.New("user is already a group member")
)

// CreateGroupRequest holds the group creation payload.
type CreateGroupRequest struct {
	GroupName    string `json:"group_name"`
	AvatarURL    string `json:"avatar_url"`
	Announcement string `json:"announcement"`
	MaxMembers   int    `json:"max_members"`
}

// GroupService defines group business operations.
type GroupService interface {
	Create(ctx context.Context, ownerID int64, req *CreateGroupRequest) (*model.GroupInfo, error)
	GetByID(ctx context.Context, groupID int64) (*model.GroupInfo, error)
	Update(ctx context.Context, userID int64, g *model.GroupInfo) error
	Dismiss(ctx context.Context, userID, groupID int64) error

	AddMember(ctx context.Context, groupID, userID int64) error
	RemoveMember(ctx context.Context, groupID, userID int64) error
	ListMembers(ctx context.Context, groupID int64) ([]*model.GroupMember, error)
}

type groupService struct {
	groupRepo   repository.GroupRepository
	sessionRepo repository.SessionRepository
}

// NewGroupService creates a GroupService.
func NewGroupService(groupRepo repository.GroupRepository, sessionRepo repository.SessionRepository) GroupService {
	return &groupService{groupRepo: groupRepo, sessionRepo: sessionRepo}
}

func (s *groupService) Create(ctx context.Context, ownerID int64, req *CreateGroupRequest) (*model.GroupInfo, error) {
	g := &model.GroupInfo{
		GroupName:    req.GroupName,
		AvatarURL:    req.AvatarURL,
		Announcement: req.Announcement,
		MaxMembers:   req.MaxMembers,
	}
	if g.MaxMembers == 0 {
		g.MaxMembers = 200
	}

	groupID, err := s.groupRepo.Create(ctx, g, ownerID)
	if err != nil {
		return nil, err
	}
	g.GroupID = groupID
	g.OwnerID = ownerID

	// Create group session
	_, _ = s.sessionRepo.GetOrCreateGroupSession(ctx, groupID)

	return g, nil
}

func (s *groupService) GetByID(ctx context.Context, groupID int64) (*model.GroupInfo, error) {
	return s.groupRepo.GetByID(ctx, groupID)
}

func (s *groupService) Update(ctx context.Context, userID int64, g *model.GroupInfo) error {
	existing, err := s.groupRepo.GetByID(ctx, g.GroupID)
	if err != nil {
		return ErrGroupNotFound
	}
	if existing.OwnerID != userID {
		return ErrNotGroupOwner
	}
	return s.groupRepo.Update(ctx, g)
}

func (s *groupService) Dismiss(ctx context.Context, userID, groupID int64) error {
	existing, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return ErrGroupNotFound
	}
	if existing.OwnerID != userID {
		return ErrNotGroupOwner
	}
	return s.groupRepo.Dismiss(ctx, groupID, userID)
}

func (s *groupService) AddMember(ctx context.Context, groupID, userID int64) error {
	isMember, _ := s.groupRepo.IsMember(ctx, groupID, userID)
	if isMember {
		return ErrAlreadyMember
	}
	return s.groupRepo.AddMember(ctx, groupID, userID, model.GroupRoleMember)
}

func (s *groupService) RemoveMember(ctx context.Context, groupID, userID int64) error {
	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

func (s *groupService) ListMembers(ctx context.Context, groupID int64) ([]*model.GroupMember, error) {
	return s.groupRepo.ListMembers(ctx, groupID)
}
