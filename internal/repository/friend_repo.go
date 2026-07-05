package repository

import (
	"context"
	"time"

	"chatgoo/internal/model"

	"gofr.dev/pkg/gofr/container"
)

// FriendRepository defines data access methods for friend relationships.
type FriendRepository interface {
	CreateGroup(ctx context.Context, g *model.FriendGroup) error
	ListGroups(ctx context.Context, userID int64) ([]*model.FriendGroup, error)
	DeleteGroup(ctx context.Context, groupID, userID int64) error

	ApplyFriend(ctx context.Context, fr *model.FriendRelation) error
	ApproveFriend(ctx context.Context, relationID int64) error
	RejectFriend(ctx context.Context, relationID int64) error
	RemoveFriend(ctx context.Context, userID, friendID int64) error
	ListFriends(ctx context.Context, userID int64) ([]*model.FriendRelation, error)
	ListPendingRequests(ctx context.Context, userID int64) ([]*model.FriendRelation, error)
	GetRelation(ctx context.Context, userID, friendID int64) (*model.FriendRelation, error)
}

type friendRepository struct {
	db container.DB
}

// NewFriendRepository creates a FriendRepository.
func NewFriendRepository(db container.DB) FriendRepository {
	return &friendRepository{db: db}
}

func (r *friendRepository) CreateGroup(ctx context.Context, g *model.FriendGroup) error {
	query := `
		INSERT INTO friend_groups (user_id, group_name)
		VALUES ($1, $2)
		RETURNING group_id, created_at
	`
	return r.db.QueryRowContext(ctx, query, g.UserID, g.GroupName).Scan(&g.GroupID, &g.CreatedAt)
}

func (r *friendRepository) ListGroups(ctx context.Context, userID int64) ([]*model.FriendGroup, error) {
	query := `SELECT group_id, user_id, group_name, created_at FROM friend_groups WHERE user_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*model.FriendGroup
	for rows.Next() {
		var g model.FriendGroup
		if err := rows.Scan(&g.GroupID, &g.UserID, &g.GroupName, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, &g)
	}
	return groups, nil
}

func (r *friendRepository) DeleteGroup(ctx context.Context, groupID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM friend_groups WHERE group_id = $1 AND user_id = $2`, groupID, userID)
	return err
}

func (r *friendRepository) ApplyFriend(ctx context.Context, fr *model.FriendRelation) error {
	query := `
		INSERT INTO friend_relations (user_id, friend_id, group_id, remark, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING relation_id, applied_at, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		fr.UserID, fr.FriendID, fr.GroupID, fr.Remark, model.FriendStatusPending,
	).Scan(&fr.RelationID, &fr.AppliedAt, &fr.CreatedAt)
}

func (r *friendRepository) ApproveFriend(ctx context.Context, relationID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE friend_relations
		SET status = $1, approved_at = $2
		WHERE relation_id = $3
	`, model.FriendStatusAccepted, time.Now(), relationID)
	return err
}

func (r *friendRepository) RejectFriend(ctx context.Context, relationID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE friend_relations SET status = $1 WHERE relation_id = $2
	`, model.FriendStatusRejected, relationID)
	return err
}

func (r *friendRepository) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM friend_relations
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
	`, userID, friendID)
	return err
}

func (r *friendRepository) ListFriends(ctx context.Context, userID int64) ([]*model.FriendRelation, error) {
	query := `
		SELECT relation_id, user_id, friend_id, group_id, remark, status, applied_at, approved_at, created_at
		FROM friend_relations
		WHERE user_id = $1 AND status = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, model.FriendStatusAccepted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []*model.FriendRelation
	for rows.Next() {
		var fr model.FriendRelation
		if err := rows.Scan(&fr.RelationID, &fr.UserID, &fr.FriendID, &fr.GroupID,
			&fr.Remark, &fr.Status, &fr.AppliedAt, &fr.ApprovedAt, &fr.CreatedAt); err != nil {
			return nil, err
		}
		relations = append(relations, &fr)
	}
	return relations, nil
}

func (r *friendRepository) ListPendingRequests(ctx context.Context, userID int64) ([]*model.FriendRelation, error) {
	query := `
		SELECT relation_id, user_id, friend_id, group_id, remark, status, applied_at, approved_at, created_at
		FROM friend_relations
		WHERE friend_id = $1 AND status = $2
		ORDER BY applied_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, model.FriendStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []*model.FriendRelation
	for rows.Next() {
		var fr model.FriendRelation
		if err := rows.Scan(&fr.RelationID, &fr.UserID, &fr.FriendID, &fr.GroupID,
			&fr.Remark, &fr.Status, &fr.AppliedAt, &fr.ApprovedAt, &fr.CreatedAt); err != nil {
			return nil, err
		}
		relations = append(relations, &fr)
	}
	return relations, nil
}

func (r *friendRepository) GetRelation(ctx context.Context, userID, friendID int64) (*model.FriendRelation, error) {
	var fr model.FriendRelation
	query := `
		SELECT relation_id, user_id, friend_id, group_id, remark, status, applied_at, approved_at, created_at
		FROM friend_relations
		WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, userID, friendID).Scan(
		&fr.RelationID, &fr.UserID, &fr.FriendID, &fr.GroupID,
		&fr.Remark, &fr.Status, &fr.AppliedAt, &fr.ApprovedAt, &fr.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &fr, nil
}
