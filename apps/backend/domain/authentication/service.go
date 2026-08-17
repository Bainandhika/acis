package authentication

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/telegramclient"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/Bainandhika/acis/apps/backend/shared/security"
	"github.com/google/uuid"
)

var (
	ErrPhoneAlreadyRegistered = errors.New("nomor telepon sudah terdaftar, silakan masuk")
	ErrPhoneNotRegistered     = errors.New("nomor telepon belum terdaftar, silakan daftar akun baru")
	ErrUsernameRequired       = errors.New("nama pengguna wajib diisi untuk pendaftaran")
	ErrInvalidPhoneFormat     = errors.New("nomor telepon harus diawali dengan +628 atau 08 (contoh: 082123456781 atau +6282123456781)")
	ErrTooManyRequests        = errors.New("terlalu banyak permintaan OTP, silakan tunggu 15 menit")
	ErrInvalidVerification    = errors.New("verifikasi kode OTP gagal atau telah kedaluwarsa")
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
	tgClient    *telegramclient.Client
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
	tgClient *telegramclient.Client,
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

func (s *authService) RequestOTP(ctx context.Context, req RequestOTPReq) (*RequestOTPResponse, error) {
	rawPhone := strings.TrimSpace(req.PhoneNumber)
	if rawPhone == "" {
		return nil, ErrInvalidPhoneFormat
	}

	if !IsValidIndonesianPhone(rawPhone) && !IsTestUser(rawPhone) {
		return nil, ErrInvalidPhoneFormat
	}

	phone := NormalizePhoneNumber(rawPhone)
	req.PhoneNumber = phone

	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		action = "login"
	}

	// 1. Validation according to action (login vs register)
	existingUser, err := s.repo.FindByPhoneNumber(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by phone: %w", err)
	}

	if action == "register" {
		username := strings.TrimSpace(req.Username)
		if username == "" {
			return nil, ErrUsernameRequired
		}
		if existingUser != nil {
			return nil, ErrPhoneAlreadyRegistered
		}
	} else if action == "login" {
		if !IsTestUser(phone) && existingUser == nil {
			return nil, ErrPhoneNotRegistered
		}
	}

	// 2. Isolated Test User Handling
	if IsTestUser(phone) {
		slog.Info("[TEST AUTH] Processing test user OTP request", slog.String("phone", phone), slog.String("action", action))
		_ = s.otpCache.StoreOTP(ctx, phone, phone, TestBypassOTP, s.otpTTL)

		return &RequestOTPResponse{
			Message:             "Kode OTP uji coba berhasil dibuat.",
			AuthSession:         fmt.Sprintf("auth_test_%s", strings.TrimPrefix(phone, "+")),
			TelegramBotUsername: s.botUsername,
			DirectSent:          true,
			IsTestUser:          true,
			TestOTP:             TestBypassOTP,
		}, nil
	}

	// 3. Check rate limit
	canReq, err := s.otpCache.CanRequestOTP(ctx, phone, phone)
	if err != nil || !canReq {
		return nil, ErrTooManyRequests
	}

	// 4. Generate 6-digit secure OTP
	otpCode, err := security.GenerateOTP()
	if err != nil {
		slog.Error("Failed to generate secure OTP", slog.Any("error", err))
		return nil, errors.New("gagal membuat kode OTP")
	}

	// 5. Store encrypted OTP in Redis
	if err := s.otpCache.StoreOTP(ctx, phone, phone, otpCode, s.otpTTL); err != nil {
		slog.Error("Failed to store encrypted OTP in Redis", slog.Any("error", err))
		return nil, errors.New("gagal memproses permintaan OTP")
	}

	// 6. Generate auth session token
	sessionID := uuid.NewString()
	authSession := fmt.Sprintf("auth_%s", sessionID[:12])
	_ = s.otpCache.StoreAuthSession(ctx, authSession, phone, phone, otpCode, s.otpTTL)

	directSent := false

	// 7. Automatic Telegram OTP Dispatch (if linked)
	var targetChatID int64
	if existingUser != nil && existingUser.TelegramChatID != nil && *existingUser.TelegramChatID != 0 {
		targetChatID = *existingUser.TelegramChatID
	}

	if targetChatID != 0 && s.tgClient != nil {
		msg := fmt.Sprintf("🔐 *Kode Verifikasi ACIS*\n\nKode OTP Anda adalah: *%s*\n\nKode ini berlaku selama 5 menit. Jangan bagikan kode ini kepada siapapun.", otpCode)
		if err := s.tgClient.SendMessage(ctx, targetChatID, msg); err == nil {
			directSent = true
			slog.Info("Telegram OTP sent automatically to chat", slog.Int64("chat_id", targetChatID), slog.String("phone", phone))
		} else {
			slog.Warn("Failed to deliver automatic Telegram OTP", slog.Any("error", err), slog.Int64("chat_id", targetChatID))
		}
	}

	slog.Info("OTP request processed successfully",
		slog.String("phone", phone),
		slog.String("action", action),
		slog.Bool("direct_sent", directSent),
		slog.String("auth_session", authSession),
	)

	return &RequestOTPResponse{
		Message:             "Kode OTP berhasil dibuat dan dikirimkan.",
		AuthSession:         authSession,
		TelegramBotUsername: s.botUsername,
		DirectSent:          directSent,
	}, nil
}

func (s *authService) VerifyOTP(ctx context.Context, req VerifyOTPReq) (*AuthResponse, error) {
	rawPhone := strings.TrimSpace(req.PhoneNumber)
	if rawPhone == "" {
		return nil, ErrInvalidPhoneFormat
	}

	phone := NormalizePhoneNumber(rawPhone)
	req.PhoneNumber = phone
	req.OTP = strings.TrimSpace(req.OTP)

	// 1. Check Test User Bypass first
	isBypass := ValidateTestBypassOTP(phone, req.OTP)
	if !isBypass {
		ok, err := s.otpCache.VerifyOTP(ctx, phone, phone, req.OTP)
		if err != nil || !ok {
			return nil, ErrInvalidVerification
		}
	}

	// 2. Find existing user by phone number
	user, err := s.repo.FindByPhoneNumber(ctx, phone)
	if err != nil {
		return nil, errors.New("failed to query user by phone")
	}

	// 3. If user does not exist, auto-register / complete registration
	if user == nil {
		userName := strings.TrimSpace(req.Username)
		if userName == "" {
			if IsTestUser(phone) {
				userName = "Admin User"
			} else {
				userName = fmt.Sprintf("User_%s", phone[len(phone)-4:])
			}
		}

		user = &User{
			ID:          uuid.NewString(),
			Username:    userName,
			PhoneNumber: phone,
			Name:        userName,
		}

		if err := s.repo.CreateUser(ctx, user); err != nil {
			slog.Error("Failed to register user", slog.Any("error", err))
			return nil, errors.New("failed to create user account")
		}
	} else if req.Username != "" && (user.Username == "" || strings.HasPrefix(user.Username, "User_")) {
		_ = s.repo.UpdateUsername(ctx, user.ID, strings.TrimSpace(req.Username))
		user.Username = strings.TrimSpace(req.Username)
	}

	// 4. Look up actual family role; default to "member" if not in a family yet
	role := "member"
	if s.roleFinder != nil {
		if foundRole, err := s.roleFinder.FindRoleByUserID(ctx, user.ID); err == nil && foundRole != "" {
			role = foundRole
		}
	}

	// 5. 15-minute short-lived access token
	accessToken, err := security.GenerateAccessToken(user.ID, role, s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// 6. Generate rotating refresh token (7 days TTL)
	refreshToken, err := cache.GenerateSecureToken(32)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	if s.tokenStore != nil {
		err = s.tokenStore.StoreRefreshToken(ctx, refreshToken, cache.RefreshSession{
			UserID: user.ID,
			Role:   role,
			Email:  user.PhoneNumber,
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
			Username:       user.Username,
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
		Email:  user.PhoneNumber,
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
			Username:       user.Username,
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
