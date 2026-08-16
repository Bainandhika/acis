package authentication

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
)

type AuthRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhoneNumber(ctx context.Context, phone string) (*User, error)
	FindByEmailAndPhone(ctx context.Context, email, phone string) (*User, error)
	FindByUserID(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateTelegramChatID(ctx context.Context, userID string, chatID int64) error
}

type authRepoImpl struct {
	db *database.AppDB
}

func NewRepository(db *database.AppDB) AuthRepository {
	return &authRepoImpl{db: db}
}

func (r *authRepoImpl) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, phone_number, name, google_id, avatar_url, telegram_chat_id, created_at, updated_at FROM users WHERE email = $1`
	var user User
	err := r.db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepoImpl) FindByPhoneNumber(ctx context.Context, phone string) (*User, error) {
	query := `SELECT id, email, phone_number, name, google_id, avatar_url, telegram_chat_id, created_at, updated_at FROM users WHERE phone_number = $1`
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

func (r *authRepoImpl) FindByEmailAndPhone(ctx context.Context, email, phone string) (*User, error) {
	query := `SELECT id, email, phone_number, name, google_id, avatar_url, telegram_chat_id, created_at, updated_at FROM users WHERE email = $1 AND phone_number = $2`
	var user User
	err := r.db.GetContext(ctx, &user, query, email, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepoImpl) FindByUserID(ctx context.Context, userID string) (*User, error) {
	query := `SELECT id, email, phone_number, name, google_id, avatar_url, telegram_chat_id, created_at, updated_at FROM users WHERE id = $1`
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
	query := `INSERT INTO users (id, email, phone_number, name, google_id, avatar_url, telegram_chat_id) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, user.ID, user.Email, user.PhoneNumber, user.Name, user.GoogleID, user.AvatarURL, user.TelegramChatID).
		Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *authRepoImpl) UpdateTelegramChatID(ctx context.Context, userID string, chatID int64) error {
	query := `UPDATE users SET telegram_chat_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	return err
}
