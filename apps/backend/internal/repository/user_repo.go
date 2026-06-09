package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
)

// UserRepository defines the contract for user data operations
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUserID(ctx context.Context, userID string) (*domain.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

// userRepo implements UserRepository using sqlx
type userRepo struct {
	db *database.AppDB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *database.AppDB) UserRepository {
	return &userRepo{db: db}
}

// FindByEmail retrieves a user by email
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, name, google_id, avatar_url, created_at, updated_at 
			  FROM users WHERE email = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)

	// Handle "not found" explicitly
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Return nil, nil instead of error
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByUserID(ctx context.Context, userID string) (*domain.User, error) {
	query := `SELECT id, email, name, google_id, avatar_url, created_at, updated_at 
			  FROM users WHERE id = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, userID)

	// Handle "not found" explicitly
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Return nil, nil instead of error
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGoogleID retrieves a user by Google ID
func (r *userRepo) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	query := `SELECT id, email, name, google_id, avatar_url, created_at, updated_at 
			  FROM users WHERE google_id = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, googleID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user into the database
func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, name, google_id, avatar_url) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`

	// NamedExecContext is safer and cleaner for structs
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Email, user.Name, user.GoogleID, user.AvatarURL).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}
