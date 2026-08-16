package authentication_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bainandhika/acis/apps/backend/domain/authentication"
	"github.com/Bainandhika/acis/apps/backend/domain/telegram"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) FindByEmail(ctx context.Context, email string) (*authentication.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authentication.User), args.Error(1)
}

func (m *MockAuthRepo) FindByPhoneNumber(ctx context.Context, phone string) (*authentication.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authentication.User), args.Error(1)
}

func (m *MockAuthRepo) FindByEmailAndPhone(ctx context.Context, email, phone string) (*authentication.User, error) {
	args := m.Called(ctx, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authentication.User), args.Error(1)
}

func (m *MockAuthRepo) FindByUserID(ctx context.Context, userID string) (*authentication.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authentication.User), args.Error(1)
}

func (m *MockAuthRepo) CreateUser(ctx context.Context, user *authentication.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthRepo) UpdateTelegramChatID(ctx context.Context, userID string, chatID int64) error {
	args := m.Called(ctx, userID, chatID)
	return args.Error(0)
}

func TestAuthService_RequestOTP_NewUserSuccess(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegram.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_123", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		Email:              "newuser@example.com",
		TelegramIdentifier: "+6281234567890",
	}

	mockRepo.On("FindByEmail", ctx, "newuser@example.com").Return(nil, nil)
	mockRepo.On("FindByPhoneNumber", ctx, "+6281234567890").Return(nil, nil)

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AuthSession)
	assert.Equal(t, "acis_bot", resp.TelegramBotUsername)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_RequestOTP_TestUserBypass(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegram.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_12345678901234567890", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		Email:              "admin@acis.test",
		TelegramIdentifier: "100000001",
	}

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.IsTestUser)
	assert.Equal(t, "123456", resp.TestOTP)

	// Verify Test Bypass
	mockRepo.On("FindByEmailAndPhone", ctx, "admin@acis.test", "100000001").Return(&authentication.User{
		ID:          "test-admin-id",
		Email:       "admin@acis.test",
		PhoneNumber: "100000001",
		Name:        "Admin User",
	}, nil)

	verifyResp, err := authSvc.VerifyOTP(ctx, authentication.VerifyOTPReq{
		Email:              "admin@acis.test",
		TelegramIdentifier: "100000001",
		OTP:                "123456",
	})

	assert.NoError(t, err)
	assert.NotNil(t, verifyResp)
	assert.NotEmpty(t, verifyResp.Token)
}

func TestAuthService_RequestOTP_EmailMismatchRejection(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegram.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_123", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		Email:              "existing@example.com",
		TelegramIdentifier: "+6289999999999", // Different phone
	}

	mockRepo.On("FindByEmail", ctx, "existing@example.com").Return(&authentication.User{
		ID:          "u-1",
		Email:       "existing@example.com",
		PhoneNumber: "+6281111111111",
	}, nil)
	mockRepo.On("FindByPhoneNumber", ctx, "+6289999999999").Return(nil, nil)

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.ErrorIs(t, err, authentication.ErrEmailMismatch)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_RequestOTP_PhoneMismatchRejection(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegram.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_123", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		Email:              "different@example.com",
		TelegramIdentifier: "+6281111111111", // Registered phone
	}

	mockRepo.On("FindByEmail", ctx, "different@example.com").Return(nil, nil)
	mockRepo.On("FindByPhoneNumber", ctx, "+6281111111111").Return(&authentication.User{
		ID:          "u-1",
		Email:       "existing@example.com",
		PhoneNumber: "+6281111111111",
	}, nil)

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.ErrorIs(t, err, authentication.ErrPhoneMismatch)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_VerifyOTP_SuccessAndTokenGeneration(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegram.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_12345678901234567890", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	email := "verify@example.com"
	phone := "+6281234567890"
	otp := "654321"

	_ = otpCache.StoreOTP(ctx, email, phone, otp, 5*time.Minute)

	mockRepo.On("FindByEmailAndPhone", ctx, email, phone).Return(&authentication.User{
		ID:          "u-100",
		Email:       email,
		PhoneNumber: phone,
		Name:        "Verify User",
	}, nil)

	resp, err := authSvc.VerifyOTP(ctx, authentication.VerifyOTPReq{
		Email:              email,
		TelegramIdentifier: phone,
		OTP:                otp,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, email, resp.User.Email)
	assert.Equal(t, phone, resp.User.PhoneNumber)

	mockRepo.AssertExpectations(t)
}
