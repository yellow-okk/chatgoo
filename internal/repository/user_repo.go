package repository

import (
	"context"

	"chatgoo/internal/model"

	"gofr.dev/pkg/gofr/container"
)

// UserRepository defines data access methods for users.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, userID int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateLastLogin(ctx context.Context, userID int64) error
	UpdateStatus(ctx context.Context, userID int64, status int16) error
	ListByIDs(ctx context.Context, ids []int64) ([]*model.User, error)
}

type userRepository struct {
	db container.DB
}

// NewUserRepository creates a UserRepository with the given database connection.
func NewUserRepository(db container.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, password_hash, nickname, avatar_url, gender, signature, region, birthday)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING user_id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		user.Username, user.PasswordHash, user.Nickname,
		user.AvatarURL, user.Gender, user.Signature,
		user.Region, user.Birthday,
	).Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByID(ctx context.Context, userID int64) (*model.User, error) {
	var u model.User
	query := `
		SELECT user_id, username, password_hash, nickname, avatar_url,
		       gender, signature, region, birthday, status, last_login_at,
		       created_at, updated_at
		FROM users WHERE user_id = $1
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&u.UserID, &u.Username, &u.PasswordHash, &u.Nickname, &u.AvatarURL,
		&u.Gender, &u.Signature, &u.Region, &u.Birthday, &u.Status, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	query := `
		SELECT user_id, username, password_hash, nickname, avatar_url,
		       gender, signature, region, birthday, status, last_login_at,
		       created_at, updated_at
		FROM users WHERE username = $1
	`
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&u.UserID, &u.Username, &u.PasswordHash, &u.Nickname, &u.AvatarURL,
		&u.Gender, &u.Signature, &u.Region, &u.Birthday, &u.Status, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET nickname = $1, avatar_url = $2, gender = $3,
		    signature = $4, region = $5, birthday = $6
		WHERE user_id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Nickname, user.AvatarURL, user.Gender,
		user.Signature, user.Region, user.Birthday, user.UserID,
	)
	return err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE user_id = $1`,
		userID,
	)
	return err
}

func (r *userRepository) UpdateStatus(ctx context.Context, userID int64, status int16) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET status = $1 WHERE user_id = $2`,
		status, userID,
	)
	return err
}

func (r *userRepository) ListByIDs(ctx context.Context, ids []int64) ([]*model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := `
		SELECT user_id, username, nickname, avatar_url, gender, signature, region, status
		FROM users WHERE user_id = ANY($1)
	`
	rows, err := r.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.UserID, &u.Username, &u.Nickname, &u.AvatarURL,
			&u.Gender, &u.Signature, &u.Region, &u.Status,
		); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}
