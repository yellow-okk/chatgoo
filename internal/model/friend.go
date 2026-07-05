package model

import "time"

// FriendGroup is a user-defined grouping for friends.
type FriendGroup struct {
	GroupID   int64     `json:"group_id" db:"group_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	GroupName string    `json:"group_name" db:"group_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FriendRelation represents a friend request or established friendship.
type FriendRelation struct {
	RelationID int64      `json:"relation_id" db:"relation_id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	FriendID   int64      `json:"friend_id" db:"friend_id"`
	GroupID    *int64     `json:"group_id" db:"group_id"`
	Remark     string     `json:"remark" db:"remark"`
	Status     int16      `json:"status" db:"status"`
	AppliedAt  time.Time  `json:"applied_at" db:"applied_at"`
	ApprovedAt *time.Time `json:"approved_at" db:"approved_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

const (
	FriendStatusPending  int16 = 0
	FriendStatusAccepted int16 = 1
	FriendStatusRejected int16 = 2
	FriendStatusBlocked  int16 = 3
)
