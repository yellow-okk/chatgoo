package model

import "time"

// User represents a user account.
type User struct {
	UserID       int64      `json:"user_id" db:"user_id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Nickname     string     `json:"nickname" db:"nickname"`
	AvatarURL    string     `json:"avatar_url" db:"avatar_url"`
	Gender       int16      `json:"gender" db:"gender"`
	Signature    string     `json:"signature" db:"signature"`
	Region       string     `json:"region" db:"region"`
	Birthday     *time.Time `json:"birthday" db:"birthday"`
	Status       int16      `json:"status" db:"status"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

const (
	UserStatusOffline int16 = 0
	UserStatusOnline  int16 = 1
	UserStatusAway    int16 = 2
)

const (
	GenderUnknown int16 = 0
	GenderMale    int16 = 1
	GenderFemale  int16 = 2
)
