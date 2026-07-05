package service

import (
	"context"
	"errors"

	"chatgoo/internal/model"
	"chatgoo/internal/repository"
)

var (
	ErrFriendSelf       = errors.New("cannot add yourself as friend")
	ErrAlreadyFriends   = errors.New("already friends")
	ErrFriendRequestExists = errors.New("friend request already exists")
	ErrRequestNotFound  = errors.New("friend request not found")
)

// ApplyFriendRequest holds the friend application payload.
type ApplyFriendRequest struct {
	FriendID int64  `json:"friend_id"`
	Remark   string `json:"remark"`
	GroupID  *int64 `json:"group_id"`
	Message  string `json:"message"`
}

// FriendService defines friend business operations.
type FriendService interface {
	Apply(ctx context.Context, userID int64, req *ApplyFriendRequest) error
	Approve(ctx context.Context, userID, relationID int64) error
	Reject(ctx context.Context, userID, relationID int64) error
	Remove(ctx context.Context, userID, friendID int64) error
	ListFriends(ctx context.Context, userID int64) ([]*model.FriendRelation, error)
	ListPendingRequests(ctx context.Context, userID int64) ([]*model.FriendRelation, error)

	CreateGroup(ctx context.Context, userID int64, groupName string) (*model.FriendGroup, error)
	ListGroups(ctx context.Context, userID int64) ([]*model.FriendGroup, error)
	DeleteGroup(ctx context.Context, userID, groupID int64) error
}

type friendService struct {
	friendRepo repository.FriendRepository
	userRepo   repository.UserRepository
}

// NewFriendService creates a FriendService.
func NewFriendService(friendRepo repository.FriendRepository, userRepo repository.UserRepository) FriendService {
	return &friendService{friendRepo: friendRepo, userRepo: userRepo}
}

func (s *friendService) Apply(ctx context.Context, userID int64, req *ApplyFriendRequest) error {
	if userID == req.FriendID {
		return ErrFriendSelf
	}

	// Check target user exists
	target, err := s.userRepo.GetByID(ctx, req.FriendID)
	if err != nil || target == nil {
		return errors.New("target user not found")
	}

	// Check existing relation
	existing, _ := s.friendRepo.GetRelation(ctx, userID, req.FriendID)
	if existing != nil {
		if existing.Status == model.FriendStatusAccepted {
			return ErrAlreadyFriends
		}
		if existing.Status == model.FriendStatusPending {
			return ErrFriendRequestExists
		}
	}

	fr := &model.FriendRelation{
		UserID:   userID,
		FriendID: req.FriendID,
		GroupID:  req.GroupID,
		Remark:   req.Remark,
	}
	return s.friendRepo.ApplyFriend(ctx, fr)
}

func (s *friendService) Approve(ctx context.Context, userID, relationID int64) error {
	return s.friendRepo.ApproveFriend(ctx, relationID)
}

func (s *friendService) Reject(ctx context.Context, userID, relationID int64) error {
	return s.friendRepo.RejectFriend(ctx, relationID)
}

func (s *friendService) Remove(ctx context.Context, userID, friendID int64) error {
	return s.friendRepo.RemoveFriend(ctx, userID, friendID)
}

func (s *friendService) ListFriends(ctx context.Context, userID int64) ([]*model.FriendRelation, error) {
	return s.friendRepo.ListFriends(ctx, userID)
}

func (s *friendService) ListPendingRequests(ctx context.Context, userID int64) ([]*model.FriendRelation, error) {
	return s.friendRepo.ListPendingRequests(ctx, userID)
}

func (s *friendService) CreateGroup(ctx context.Context, userID int64, groupName string) (*model.FriendGroup, error) {
	g := &model.FriendGroup{
		UserID:    userID,
		GroupName: groupName,
	}
	if err := s.friendRepo.CreateGroup(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *friendService) ListGroups(ctx context.Context, userID int64) ([]*model.FriendGroup, error) {
	return s.friendRepo.ListGroups(ctx, userID)
}

func (s *friendService) DeleteGroup(ctx context.Context, userID, groupID int64) error {
	return s.friendRepo.DeleteGroup(ctx, groupID, userID)
}
