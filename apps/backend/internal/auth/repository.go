package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUserID(ctx context.Context, userID string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
}

type authRepoImpl struct {
	db *database.AppDB
}

func NewRepository(db *database.AppDB) Repository {
	return &authRepoImpl{db: db}
}

func (r *authRepoImpl) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, name, google_id, avatar_url, created_at, updated_at FROM users WHERE email = $1`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepoImpl) FindByUserID(ctx context.Context, userID string) (*domain.User, error) {
	query := `SELECT id, email, name, google_id, avatar_url, created_at, updated_at FROM users WHERE id = $1`
	var user domain.User
	err := r.db.GetContext(ctx, &user, query, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepoImpl) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, name, google_id, avatar_url) VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, user.ID, user.Email, user.Name, user.GoogleID, user.AvatarURL).Scan(&user.CreatedAt, &user.UpdatedAt)
}
