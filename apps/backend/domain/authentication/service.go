package authentication

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/Bainandhika/acis/apps/backend/shared/security"
	"github.com/google/uuid"
)

// RoleFinder looks up a user's family role. Implemented by the family repository adapter in bootstrap.
type RoleFinder interface {
	FindRoleByUserID(ctx context.Context, userID string) (string, error)
}

type AuthService interface {
	RequestOTP(ctx context.Context, req RequestOTPReq) error
	VerifyOTP(ctx context.Context, req VerifyOTPReq) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	repo        AuthRepository
	roleFinder  RoleFinder
	outboxRepo  notification.OutboxRepository
	otpCache    *cache.OTPCache
	tokenStore  *cache.RefreshTokenStore
	db          *database.AppDB
	jwtSecret   string
	otpTTL      time.Duration
}

func NewService(repo AuthRepository, roleFinder RoleFinder, outboxRepo notification.OutboxRepository, otpCache *cache.OTPCache, tokenStore *cache.RefreshTokenStore, db *database.AppDB, jwtSecret string, otpTTL time.Duration) AuthService {
	if otpTTL <= 0 {
		otpTTL = 5 * time.Minute
	}
	return &authService{
		repo:        repo,
		roleFinder:  roleFinder,
		outboxRepo:  outboxRepo,
		otpCache:    otpCache,
		tokenStore:  tokenStore,
		db:          db,
		jwtSecret:   jwtSecret,
		otpTTL:      otpTTL,
	}
}

func (s *authService) RequestOTP(ctx context.Context, req RequestOTPReq) error {
	canReq, err := s.otpCache.CanRequestOTP(ctx, req.Email)
	if err != nil || !canReq {
		return errors.New("too many OTP requests. Please wait 15 minutes")
	}

	otpCode, err := security.GenerateOTP()
	if err != nil {
		slog.Error("Failed to generate secure OTP", slog.Any("error", err))
		return errors.New("failed to generate OTP")
	}

	if err := s.otpCache.StoreOTP(ctx, req.Email, otpCode, s.otpTTL); err != nil {
		slog.Error("Failed to store encrypted OTP in Redis", slog.Any("error", err))
		return errors.New("failed to process OTP request")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("Failed to start db transaction for OTP outbox", slog.Any("error", err))
		return errors.New("database transaction error")
	}
	defer tx.Rollback()

	payload := map[string]string{
		"code":  otpCode,
		"email": req.Email,
	}

	if err := s.outboxRepo.EnqueueTx(ctx, tx, "email_otp", req.Email, payload); err != nil {
		slog.Error("Failed to enqueue OTP to notification outbox", slog.Any("error", err))
		return errors.New("failed to queue notification")
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit outbox transaction", slog.Any("error", err))
		return errors.New("failed to commit transaction")
	}

	slog.Info("OTP requested & enqueued to outbox", slog.String("email", req.Email))
	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, req VerifyOTPReq) (*AuthResponse, error) {
	ok, err := s.otpCache.VerifyOTP(ctx, req.Email, req.OTP)
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
			slog.Error("Failed to auto-register user", slog.Any("error", err))
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

	// 15-minute short-lived access token
	accessToken, err := security.GenerateAccessToken(user.ID, role, s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate single-use rotating refresh token and store in Redis (7 days TTL)
	refreshToken, err := cache.GenerateSecureToken(32)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	if s.tokenStore != nil {
		err = s.tokenStore.StoreRefreshToken(ctx, refreshToken, cache.RefreshSession{
			UserID: user.ID,
			Role:   role,
			Email:  user.Email,
		}, 7*24*time.Hour)
		if err != nil {
			slog.Error("Failed to store refresh token in Redis", slog.Any("error", err))
			return nil, errors.New("failed to establish session")
		}
	}

	return &AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      role,
			AvatarURL: user.AvatarURL,
		},
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	if refreshToken == "" || s.tokenStore == nil {
		return nil, errors.New("missing or invalid refresh token")
	}

	// Single-use token rotation: retrieve and atomically delete old refresh token
	session, err := s.tokenStore.GetAndRevokeRefreshToken(ctx, refreshToken)
	if err != nil {
		slog.Warn("Refresh token invalid or expired", slog.Any("error", err))
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.repo.FindByUserID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Fetch up-to-date role
	role := session.Role
	if s.roleFinder != nil {
		if foundRole, err := s.roleFinder.FindRoleByUserID(ctx, user.ID); err == nil && foundRole != "" {
			role = foundRole
		}
	}

	// Generate new 15-minute access token
	newAccessToken, err := security.GenerateAccessToken(user.ID, role, s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate new access token")
	}

	// Generate new single-use refresh token
	newRefreshToken, err := cache.GenerateSecureToken(32)
	if err != nil {
		return nil, errors.New("failed to generate new refresh token")
	}

	err = s.tokenStore.StoreRefreshToken(ctx, newRefreshToken, cache.RefreshSession{
		UserID: user.ID,
		Role:   role,
		Email:  user.Email,
	}, 7*24*time.Hour)
	if err != nil {
		slog.Error("Failed to store rotated refresh token in Redis", slog.Any("error", err))
		return nil, errors.New("failed to refresh session")
	}

	slog.Info("Access and refresh tokens successfully rotated", slog.String("user_id", user.ID))

	return &AuthResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      role,
			AvatarURL: user.AvatarURL,
		},
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken != "" && s.tokenStore != nil {
		if err := s.tokenStore.RevokeRefreshToken(ctx, refreshToken); err != nil {
			slog.Warn("Failed to revoke refresh token from Redis", slog.Any("error", err))
		}
	}
	return nil
}
