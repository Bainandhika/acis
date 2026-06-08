package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/google/uuid"
)

type AuthRepository interface {
	SaveOTP(ctx context.Context, email, hashedCode string, expiresAt time.Time) error
	GetLatestActiveOTP(ctx context.Context, email string) (*domain.OTPCode, error)
	MarkOTPAsUsed(ctx context.Context, otpID string) error
}

type authRepo struct {
	db *database.AppDB
}

func NewAuthRepository(db *database.AppDB) AuthRepository {
	return &authRepo{db: db}
}

func (r *authRepo) SaveOTP(ctx context.Context, email, hashedCode string, expiresAt time.Time) error {
	id := uuid.New().String()

	query := `INSERT INTO otp_codes (id, email, code_hash, expires_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, id, email, hashedCode, expiresAt)
	
	return err
}

func (r *authRepo) GetLatestActiveOTP(ctx context.Context, email string) (*domain.OTPCode, error) {
	query := `SELECT id, email, code_hash, expires_at, is_used, created_at 
			  FROM otp_codes 
			  WHERE email = $1 AND is_used = false AND expires_at > CURRENT_TIMESTAMP 
			  ORDER BY created_at DESC LIMIT 1`

	var otp domain.OTPCode
	err := r.db.GetContext(ctx, &otp, query, email)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No active OTP found
	}
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *authRepo) MarkOTPAsUsed(ctx context.Context, otpID string) error {
	query := `UPDATE otp_codes SET is_used = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, otpID)
	return err
}
