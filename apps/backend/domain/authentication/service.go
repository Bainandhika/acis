package authentication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/Bainandhika/acis/apps/backend/shared/security"
	"github.com/google/uuid"
	zerolog "github.com/rs/zerolog/log"
)

// RoleFinder looks up a user's family role. Implemented by the family repository adapter in bootstrap.
type RoleFinder interface {
	FindRoleByUserID(ctx context.Context, userID string) (string, error)
}

type AuthService interface {
	RequestOTP(ctx context.Context, req RequestOTPReq) error
	VerifyOTP(ctx context.Context, req VerifyOTPReq) (*AuthResponse, error)
}

type authService struct {
	repo       AuthRepository
	roleFinder RoleFinder
	outboxRepo notification.OutboxRepository
	otpCache   *cache.OTPCache
	db         *database.AppDB
	jwtSecret  string
}

func NewService(repo AuthRepository, roleFinder RoleFinder, outboxRepo notification.OutboxRepository, otpCache *cache.OTPCache, db *database.AppDB, jwtSecret string) AuthService {
	return &authService{
		repo:       repo,
		roleFinder: roleFinder,
		outboxRepo: outboxRepo,
		otpCache:   otpCache,
		db:         db,
		jwtSecret:  jwtSecret,
	}
}

func (s *authService) RequestOTP(ctx context.Context, req RequestOTPReq) error {
	if !s.otpCache.CanRequestOTP(req.Email) {
		return errors.New("too many OTP requests. Please wait 15 minutes")
	}

	otpCode, err := security.GenerateOTP()
	if err != nil {
		zerolog.Error().Err(err).Msg("Failed to generate secure OTP")
		return errors.New("failed to generate OTP")
	}

	s.otpCache.StoreOTP(req.Email, otpCode, 5*time.Minute)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		zerolog.Error().Err(err).Msg("Failed to start db transaction for OTP outbox")
		return errors.New("database transaction error")
	}
	defer tx.Rollback()

	payload := map[string]string{
		"code":  otpCode,
		"email": req.Email,
	}

	if err := s.outboxRepo.EnqueueTx(ctx, tx, "email_otp", req.Email, payload); err != nil {
		zerolog.Error().Err(err).Msg("Failed to enqueue OTP to notification outbox")
		return errors.New("failed to queue notification")
	}

	if err := tx.Commit(); err != nil {
		zerolog.Error().Err(err).Msg("Failed to commit outbox transaction")
		return errors.New("failed to commit transaction")
	}

	zerolog.Info().Str("email", req.Email).Msg("OTP requested & enqueued to outbox")
	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, req VerifyOTPReq) (*AuthResponse, error) {
	ok, err := s.otpCache.VerifyOTP(req.Email, req.OTP)
	if err != nil || !ok {
		return nil, fmt.Errorf("otp verification failed: %w", err)
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("failed to query user")
	}

	if user == nil {
		user = &User{
			ID:    uuid.NewString(),
			Email: req.Email,
			Name:  req.Email,
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			zerolog.Error().Err(err).Msg("Failed to auto-register user")
			return nil, errors.New("failed to create user")
		}
	}

	// Look up the user's actual family role; default to "member" if not in a family yet
	role := "member"
	if s.roleFinder != nil {
		if foundRole, err := s.roleFinder.FindRoleByUserID(ctx, user.ID); err == nil && foundRole != "" {
			role = foundRole
		}
	}

	token, err := security.GenerateToken(user.ID, role, s.jwtSecret, 24)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &AuthResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      role,
			AvatarURL: user.AvatarURL,
		},
	}, nil
}
