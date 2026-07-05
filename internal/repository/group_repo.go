package repository

import (
	"context"

	"chatgoo/internal/model"

	"gofr.dev/pkg/gofr/container"
)

// GroupRepository defines data access methods for groups.
type GroupRepository interface {
	Create(ctx context.Context, g *model.GroupInfo, ownerID int64) (int64, error)
	GetByID(ctx context.Context, groupID int64) (*model.GroupInfo, error)
	Update(ctx context.Context, g *model.GroupInfo) error
	Dismiss(ctx context.Context, groupID, ownerID int64) error

	AddMember(ctx context.Context, groupID, userID int64, role int16) error
	RemoveMember(ctx context.Context, groupID, userID int64) error
	UpdateMemberRole(ctx context.Context, groupID, userID int64, role int16) error
	ListMembers(ctx context.Context, groupID int64) ([]*model.GroupMember, error)
	IsMember(ctx context.Context, groupID, userID int64) (bool, error)
}

type groupRepository struct {
	db container.DB
}

// NewGroupRepository creates a GroupRepository.
func NewGroupRepository(db container.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(ctx context.Context, g *model.GroupInfo, ownerID int64) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO group_info (group_name, owner_id, avatar_url, announcement, max_members)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING group_id, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query,
		g.GroupName, ownerID, g.AvatarURL, g.Announcement, g.MaxMembers,
	).Scan(&g.GroupID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, $3)
	`, g.GroupID, ownerID, model.GroupRoleOwner)
	if err != nil {
		return 0, err
	}

	return g.GroupID, tx.Commit()
}

func (r *groupRepository) GetByID(ctx context.Context, groupID int64) (*model.GroupInfo, error) {
	var g model.GroupInfo
	query := `
		SELECT group_id, group_name, owner_id, avatar_url, announcement, max_members, created_at, updated_at
		FROM group_info WHERE group_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, groupID).Scan(
		&g.GroupID, &g.GroupName, &g.OwnerID, &g.AvatarURL,
		&g.Announcement, &g.MaxMembers, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *groupRepository) Update(ctx context.Context, g *model.GroupInfo) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE group_info
		SET group_name = $1, avatar_url = $2, announcement = $3, max_members = $4
		WHERE group_id = $5
	`, g.GroupName, g.AvatarURL, g.Announcement, g.MaxMembers, g.GroupID)
	return err
}

func (r *groupRepository) Dismiss(ctx context.Context, groupID, ownerID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM group_info WHERE group_id = $1 AND owner_id = $2`, groupID, ownerID)
	return err
}

func (r *groupRepository) AddMember(ctx context.Context, groupID, userID int64, role int16) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, groupID, userID, role)
	return err
}

func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`, groupID, userID)
	return err
}

func (r *groupRepository) UpdateMemberRole(ctx context.Context, groupID, userID int64, role int16) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE group_members SET role = $1 WHERE group_id = $2 AND user_id = $3`, role, groupID, userID)
	return err
}

func (r *groupRepository) ListMembers(ctx context.Context, groupID int64) ([]*model.GroupMember, error) {
	query := `SELECT id, group_id, user_id, role, join_at FROM group_members WHERE group_id = $1 ORDER BY join_at`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.GroupMember
	for rows.Next() {
		var m model.GroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.Role, &m.JoinAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *groupRepository) IsMember(ctx context.Context, groupID, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)`,
		groupID, userID).Scan(&exists)
	return exists, err
}
