package authentication

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/jmoiron/sqlx"
)

type AuthRepository interface {
	FindByUserID(ctx context.Context, userID, email string) (*User, error)
	ProvisionUser(ctx context.Context, userID, email string, user *User) (*User, error)
	GetUserMemberships(ctx context.Context, userID, email string) ([]FamilyMembership, error)
	UpdateTelegramChatID(ctx context.Context, userID string, chatID int64) error
	UpdateTelegramChatIDByUserID(ctx context.Context, userID, email string, chatID int64) error
}

type authRepoImpl struct {
	db *database.AppDB
}

func NewRepository(db *database.AppDB) AuthRepository {
	return &authRepoImpl{db: db}
}

func (r *authRepoImpl) FindByUserID(ctx context.Context, userID, email string) (*User, error) {
	var user User
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `SELECT id, username, name, avatar_url, telegram_chat_id, created_at, updated_at FROM users WHERE id = $1`
		err := tx.GetContext(ctx, &user, query, userID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if user.ID == "" {
		return nil, nil
	}
	return &user, nil
}

func (r *authRepoImpl) ProvisionUser(ctx context.Context, userID, email string, user *User) (*User, error) {
	var provisioned User
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		// username fallback = 'user_' || left($1, 8)
		query := `
			INSERT INTO users (id, name, username, avatar_url)
			VALUES ($1, $2, COALESCE(NULLIF(TRIM($3), ''), 'user_' || LEFT($1, 8)), $4)
			ON CONFLICT (id) DO NOTHING
		`
		if _, err := tx.ExecContext(ctx, query, user.ID, user.Name, user.Username, user.AvatarURL); err != nil {
			return err
		}

		selectQuery := `SELECT id, username, name, avatar_url, telegram_chat_id, created_at, updated_at FROM users WHERE id = $1`
		return tx.GetContext(ctx, &provisioned, selectQuery, user.ID)
	})
	if err != nil {
		return nil, err
	}
	return &provisioned, nil
}

func (r *authRepoImpl) GetUserMemberships(ctx context.Context, userID, email string) ([]FamilyMembership, error) {
	var memberships []FamilyMembership
	err := r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `
			SELECT fm.family_id, f.name AS family_name, fm.role
			FROM family_members fm
			JOIN families f ON fm.family_id = f.id
			WHERE fm.user_id = $1
			ORDER BY fm.joined_at ASC
		`
		return tx.SelectContext(ctx, &memberships, query, userID)
	})
	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *authRepoImpl) UpdateTelegramChatID(ctx context.Context, userID string, chatID int64) error {
	// Used by internal bot endpoint with AdminDB (or direct exec)
	query := `UPDATE users SET telegram_chat_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.AdminDB().ExecContext(ctx, query, chatID, userID)
	return err
}

func (r *authRepoImpl) UpdateTelegramChatIDByUserID(ctx context.Context, userID, email string, chatID int64) error {
	return r.db.WithUserContext(ctx, func(tx *sqlx.Tx) error {
		query := `UPDATE users SET telegram_chat_id = $1, updated_at = NOW() WHERE id = $2`
		_, err := tx.ExecContext(ctx, query, chatID, userID)
		return err
	})
}
