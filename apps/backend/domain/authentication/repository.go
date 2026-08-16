package authentication

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
)

type AuthRepository interface {
	FindByPhoneNumber(ctx context.Context, phone string) (*User, error)
	FindByUserID(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateTelegramChatID(ctx context.Context, userID string, chatID int64) error
	UpdateUsername(ctx context.Context, userID string, username string) error
}

type authRepoImpl struct {
	db *database.AppDB
}

func NewRepository(db *database.AppDB) AuthRepository {
	return &authRepoImpl{db: db}
}

func (r *authRepoImpl) FindByPhoneNumber(ctx context.Context, phone string) (*User, error) {
	query := `SELECT id, COALESCE(username, name) as username, phone_number, name, avatar_url, telegram_chat_id, created_at, updated_at FROM users WHERE phone_number = $1`
	var user User
	err := r.db.GetContext(ctx, &user, query, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepoImpl) FindByUserID(ctx context.Context, userID string) (*User, error) {
	query := `SELECT id, COALESCE(username, name) as username, phone_number, name, avatar_url, telegram_chat_id, created_at, updated_at FROM users WHERE id = $1`
	var user User
	err := r.db.GetContext(ctx, &user, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepoImpl) CreateUser(ctx context.Context, user *User) error {
	query := `INSERT INTO users (id, username, phone_number, name, avatar_url, telegram_chat_id) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, user.ID, user.Username, user.PhoneNumber, user.Name, user.AvatarURL, user.TelegramChatID).
		Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *authRepoImpl) UpdateTelegramChatID(ctx context.Context, userID string, chatID int64) error {
	query := `UPDATE users SET telegram_chat_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	return err
}

func (r *authRepoImpl) UpdateUsername(ctx context.Context, userID string, username string) error {
	query := `UPDATE users SET username = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, username, userID)
	return err
}
