package model

import "time"

// GroupInfo represents a chat group.
type GroupInfo struct {
	GroupID      int64     `json:"group_id" db:"group_id"`
	GroupName    string    `json:"group_name" db:"group_name"`
	OwnerID      int64     `json:"owner_id" db:"owner_id"`
	AvatarURL    string    `json:"avatar_url" db:"avatar_url"`
	Announcement string    `json:"announcement" db:"announcement"`
	MaxMembers   int       `json:"max_members" db:"max_members"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// GroupMember represents a user's membership in a group.
type GroupMember struct {
	ID      int64     `json:"id" db:"id"`
	GroupID int64     `json:"group_id" db:"group_id"`
	UserID  int64     `json:"user_id" db:"user_id"`
	Role    int16     `json:"role" db:"role"`
	JoinAt  time.Time `json:"join_at" db:"join_at"`
}

const (
	GroupRoleMember int16 = 0
	GroupRoleAdmin  int16 = 1
	GroupRoleOwner  int16 = 2
)
