package authentication

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Bainandhika/acis/apps/backend/domain/telegram"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/Bainandhika/acis/apps/backend/shared/security"
	"github.com/google/uuid"
)

var (
	ErrEmailMismatch       = errors.New("email is already registered with a different Telegram identifier/phone")
	ErrPhoneMismatch       = errors.New("telegram identifier/phone is already registered with a different email")
	ErrAccountConflict     = errors.New("email and Telegram identifier belong to different accounts")
	ErrTooManyRequests     = errors.New("too many OTP requests. Please wait 15 minutes")
	ErrInvalidVerification = errors.New("otp verification failed")
)

// RoleFinder looks up a user's family role. Implemented by the family repository adapter in bootstrap.
type RoleFinder interface {
	FindRoleByUserID(ctx context.Context, userID string) (string, error)
}

type AuthService interface {
	RequestOTP(ctx context.Context, req RequestOTPReq) (*RequestOTPResponse, error)
	VerifyOTP(ctx context.Context, req VerifyOTPReq) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetOTPCache() *cache.OTPCache
}

type authService struct {
	repo        AuthRepository
	roleFinder  RoleFinder
	outboxRepo  notification.OutboxRepository
	otpCache    *cache.OTPCache
	tokenStore  *cache.RefreshTokenStore
	tgClient    *telegram.Client
	db          *database.AppDB
	jwtSecret   string
	botUsername string
	otpTTL      time.Duration
}

func NewService(
	repo AuthRepository,
	roleFinder RoleFinder,
	outboxRepo notification.OutboxRepository,
	otpCache *cache.OTPCache,
	tokenStore *cache.RefreshTokenStore,
	tgClient *telegram.Client,
	db *database.AppDB,
	jwtSecret string,
	botUsername string,
	otpTTL time.Duration,
) AuthService {
	if otpTTL <= 0 {
		otpTTL = 5 * time.Minute
	}
	return &authService{
		repo:        repo,
		roleFinder:  roleFinder,
		outboxRepo:  outboxRepo,
		otpCache:    otpCache,
		tokenStore:  tokenStore,
		tgClient:    tgClient,
		db:          db,
		jwtSecret:   jwtSecret,
		botUsername: botUsername,
		otpTTL:      otpTTL,
	}
}

func (s *authService) GetOTPCache() *cache.OTPCache {
	return s.otpCache
}

func resolveIdentifier(req *RequestOTPReq) string {
	identifier := strings.TrimSpace(req.TelegramIdentifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.PhoneNumber)
	}
	if identifier == "" {
		identifier = "user_tg"
	}
	req.PhoneNumber = identifier
	req.TelegramIdentifier = identifier
	return identifier
}

func resolveVerifyIdentifier(req *VerifyOTPReq) string {
	identifier := strings.TrimSpace(req.TelegramIdentifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.PhoneNumber)
	}
	if identifier == "" {
		identifier = "user_tg"
	}
	req.PhoneNumber = identifier
	req.TelegramIdentifier = identifier
	return identifier
}

func (s *authService) RequestOTP(ctx context.Context, req RequestOTPReq) (*RequestOTPResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	identifier := resolveIdentifier(&req)

	// Isolated Test User Handling
	if IsTestUser(req.Email) {
		slog.Info("[TEST AUTH] Processing test user OTP request", slog.String("email", req.Email))
		// Store test OTP in cache as well
		_ = s.otpCache.StoreOTP(ctx, req.Email, identifier, TestBypassOTP, s.otpTTL)

		return &RequestOTPResponse{
			Message:             "Test OTP generated successfully. Use bypass OTP 123456.",
			AuthSession:         fmt.Sprintf("auth_test_%s", strings.Split(req.Email, "@")[0]),
			TelegramBotUsername: s.botUsername,
			DirectSent:          true,
			IsTestUser:          true,
			TestOTP:             TestBypassOTP,
		}, nil
	}

	// 1. Enforce strict uniqueness & anti-reuse rules
	userByEmail, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	userByPhone, err := s.repo.FindByPhoneNumber(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by phone/identifier: %w", err)
	}

	if userByEmail != nil && userByPhone != nil {
		if userByEmail.ID != userByPhone.ID {
			return nil, ErrAccountConflict
		}
	} else if userByEmail != nil {
		if userByEmail.PhoneNumber != identifier {
			return nil, ErrEmailMismatch
		}
	} else if userByPhone != nil {
		if userByPhone.Email != req.Email {
			return nil, ErrPhoneMismatch
		}
	}

	// 2. Check rate limit
	canReq, err := s.otpCache.CanRequestOTP(ctx, req.Email, identifier)
	if err != nil || !canReq {
		return nil, ErrTooManyRequests
	}

	// 3. Generate 6-digit OTP
	otpCode, err := security.GenerateOTP()
	if err != nil {
		slog.Error("Failed to generate secure OTP", slog.Any("error", err))
		return nil, errors.New("failed to generate OTP")
	}

	// 4. Store encrypted OTP in Redis
	if err := s.otpCache.StoreOTP(ctx, req.Email, identifier, otpCode, s.otpTTL); err != nil {
		slog.Error("Failed to store encrypted OTP in Redis", slog.Any("error", err))
		return nil, errors.New("failed to process OTP request")
	}

	// 5. Generate auth session token
	sessionID := uuid.NewString()
	authSession := fmt.Sprintf("auth_%s", sessionID[:12])
	_ = s.otpCache.StoreAuthSession(ctx, authSession, req.Email, identifier, otpCode, s.otpTTL)

	directSent := false
	var existingUser *User
	if userByEmail != nil {
		existingUser = userByEmail
	} else if userByPhone != nil {
		existingUser = userByPhone
	}

	// 6. Automatic Telegram OTP Dispatch (Without /start required)
	// Try delivering to existing user's linked Telegram Chat ID OR if identifier is numeric Chat ID
	var targetChatID int64
	if existingUser != nil && existingUser.TelegramChatID != nil && *existingUser.TelegramChatID != 0 {
		targetChatID = *existingUser.TelegramChatID
	} else if parsedChatID, parseErr := strconv.ParseInt(identifier, 10, 64); parseErr == nil && parsedChatID != 0 {
		targetChatID = parsedChatID
	}

	if targetChatID != 0 && s.tgClient != nil {
		msg := fmt.Sprintf("🔐 *ACIS Verification Code*\n\nYour OTP is: *%s*\n\nThis code is valid for 5 minutes. Do not share this code with anyone.", otpCode)
		if err := s.tgClient.SendMessage(ctx, targetChatID, msg); err == nil {
			directSent = true
			slog.Info("Telegram OTP sent automatically to chat", slog.Int64("chat_id", targetChatID), slog.String("email", req.Email))
		} else {
			slog.Warn("Failed to deliver automatic Telegram OTP", slog.Any("error", err), slog.Int64("chat_id", targetChatID))
		}
	}

	slog.Info("Telegram OTP processed successfully",
		slog.String("email", req.Email),
		slog.String("identifier", identifier),
		slog.Bool("direct_sent", directSent),
		slog.String("auth_session", authSession),
	)

	return &RequestOTPResponse{
		Message:             "Telegram OTP has been generated and sent automatically.",
		AuthSession:         authSession,
		TelegramBotUsername: s.botUsername,
		DirectSent:          directSent,
	}, nil
}

func (s *authService) VerifyOTP(ctx context.Context, req VerifyOTPReq) (*AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	identifier := resolveVerifyIdentifier(&req)
	req.OTP = strings.TrimSpace(req.OTP)

	// Check Test User Bypass first
	isBypass := ValidateTestBypassOTP(req.Email, req.OTP)
	if !isBypass {
		ok, err := s.otpCache.VerifyOTP(ctx, req.Email, identifier, req.OTP)
		if err != nil || !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidVerification, err)
		}
	}

	user, err := s.repo.FindByEmailAndPhone(ctx, req.Email, identifier)
	if err != nil {
		return nil, errors.New("failed to query user")
	}

	if user == nil {
		// New user registration: double check uniqueness constraints before creating
		existingEmail, _ := s.repo.FindByEmail(ctx, req.Email)
		if existingEmail != nil {
			return nil, ErrEmailMismatch
		}
		existingPhone, _ := s.repo.FindByPhoneNumber(ctx, identifier)
		if existingPhone != nil {
			return nil, ErrPhoneMismatch
		}

		userName := strings.Split(req.Email, "@")[0]
		user = &User{
			ID:          uuid.NewString(),
			Email:       req.Email,
			PhoneNumber: identifier,
			Name:        userName,
		}

		// If identifier is numeric, auto-associate TelegramChatID
		if parsedChatID, err := strconv.ParseInt(identifier, 10, 64); err == nil && parsedChatID != 0 {
			user.TelegramChatID = &parsedChatID
		}

		if err := s.repo.CreateUser(ctx, user); err != nil {
			slog.Error("Failed to auto-register user", slog.Any("error", err))
			return nil, errors.New("failed to create user")
		}
	}

	// Look up actual family role; default to "member" if not in a family yet
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

	// Generate rotating refresh token (7 days TTL)
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
			ID:             user.ID,
			Email:          user.Email,
			PhoneNumber:    user.PhoneNumber,
			Name:           user.Name,
			Role:           role,
			AvatarURL:      user.AvatarURL,
			TelegramChatID: user.TelegramChatID,
		},
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	if refreshToken == "" || s.tokenStore == nil {
		return nil, errors.New("missing or invalid refresh token")
	}

	session, err := s.tokenStore.GetAndRevokeRefreshToken(ctx, refreshToken)
	if err != nil {
		slog.Warn("Refresh token invalid or expired", slog.Any("error", err))
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.repo.FindByUserID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	role := session.Role
	if s.roleFinder != nil {
		if foundRole, err := s.roleFinder.FindRoleByUserID(ctx, user.ID); err == nil && foundRole != "" {
			role = foundRole
		}
	}

	newAccessToken, err := security.GenerateAccessToken(user.ID, role, s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate new access token")
	}

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

	return &AuthResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		User: UserResponse{
			ID:             user.ID,
			Email:          user.Email,
			PhoneNumber:    user.PhoneNumber,
			Name:           user.Name,
			Role:           role,
			AvatarURL:      user.AvatarURL,
			TelegramChatID: user.TelegramChatID,
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
