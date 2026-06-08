package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/Bainandhika/acis/apps/backend/pkg/auth"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type AuthService interface {
	RequestOTP(ctx context.Context, req dto.RequestOTPRequest) error
	VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.AuthResponse, error)
}

type authService struct {
	authRepo  repository.AuthRepository
	userRepo  repository.UserRepository
	jwtSecret string

	// In-memory rate limiter for MVP
	rateLimitMap map[string]time.Time
	mu           sync.Mutex
}

func NewAuthService(authRepo repository.AuthRepository, userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{
		authRepo:     authRepo,
		userRepo:     userRepo,
		jwtSecret:    jwtSecret,
		rateLimitMap: make(map[string]time.Time),
	}
}

// RequestOTP handles OTP generation and "sending"
func (s *authService) RequestOTP(ctx context.Context, req dto.RequestOTPRequest) error {
	// 1. Rate Limiting Check (Max 1 request per 60 seconds per email)
	s.mu.Lock()
	if lastReq, exists := s.rateLimitMap[req.Email]; exists {
		if time.Since(lastReq) < 60*time.Second {
			s.mu.Unlock()
			return errors.New("too many requests. please wait 60 seconds")
		}
	}
	s.rateLimitMap[req.Email] = time.Now()
	s.mu.Unlock()

	// 2. Generate Cryptographically Secure OTP
	otpCode, err := auth.GenerateOTP()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate OTP")
		return errors.New("failed to generate OTP")
	}

	// 3. Hash OTP before saving (OWASP A02)
	hashedOTP, err := auth.HashOTP(otpCode)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash OTP")
		return errors.New("failed to process OTP")
	}

	// 4. Save to Database
	expiresAt := time.Now().Add(auth.OTPExpiry)
	if err := s.authRepo.SaveOTP(ctx, req.Email, hashedOTP, expiresAt); err != nil {
		log.Error().Err(err).Msg("Failed to save OTP to DB")
		return errors.New("failed to save OTP")
	}

	// 5. Mock Email Sending (For MVP, we log it. In prod, use Resend/SendGrid)
	// TODO: Integrate with Resend API here
	log.Info().Str("email", req.Email).Str("otp_code", otpCode).Msg("OTP Generated (Mock Email Send)")

	return nil
}

// VerifyOTP validates the OTP and returns a JWT
func (s *authService) VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.AuthResponse, error) {
	// 1. Get the latest active OTP from DB
	otpRecord, err := s.authRepo.GetLatestActiveOTP(ctx, req.Email)
	if err != nil {
		log.Error().Err(err).Msg("DB error while fetching OTP")
		return nil, errors.New("verification failed")
	}
	if otpRecord == nil {
		return nil, errors.New("no active OTP found or OTP expired")
	}

	// 2. Verify the OTP code against the hash
	if !auth.VerifyOTP(req.Code, otpRecord.CodeHash) {
		return nil, errors.New("invalid OTP code")
	}

	// 3. Mark OTP as used (One-time use)
	if err := s.authRepo.MarkOTPAsUsed(ctx, otpRecord.ID); err != nil {
		log.Error().Err(err).Msg("Failed to mark OTP as used")
		// Continue anyway, but log it
	}

	// 4. Find or Create User (Frictionless Onboarding)
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("failed to fetch user")
	}

	if user == nil {
		// Auto-register new user for MVP
		log.Info().Str("email", req.Email).Msg("New user detected, auto-registering...")
		
		// Generate UUID for new user
		userID, err := uuid.NewUUID()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate UUID for new user")
			return nil, errors.New("failed to create user")
		}
		
		user = &domain.User{
			ID:    userID.String(),
			Email: req.Email,
			Name:  req.Email, // Default name to email, user can update later
		}
		
		// Create user in database
		if err := s.userRepo.Create(ctx, user); err != nil {
			log.Error().Err(err).Msg("Failed to create new user in DB")
			return nil, errors.New("failed to register user")
		}
		
		log.Info().Str("user_id", user.ID).Msg("New user registered successfully")
	}

	// 5. Determine Role (For MVP, check if user is admin of any family, else member)
	// TODO: Implement proper role checking logic later. Default to 'member'.
	role := "member"

	// 6. Generate JWT
	token, err := auth.GenerateToken(user.ID, role, s.jwtSecret, 24) // 24 hours expiry
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT")
		return nil, errors.New("failed to generate token")
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  role,
		},
	}, nil
}
