package authentication_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bainandhika/acis/apps/backend/domain/authentication"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/telegramclient"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) FindByPhoneNumber(ctx context.Context, phone string) (*authentication.User, error) {
	args := m.Called(ctx, phone)
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

func (m *MockAuthRepo) UpdateUsername(ctx context.Context, userID string, username string) error {
	args := m.Called(ctx, userID, username)
	return args.Error(0)
}

func TestAuthService_RequestOTP_SignIn_Success(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegramclient.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_123", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		PhoneNumber: "081234567890",
		Action:      "login",
	}

	mockRepo.On("FindByPhoneNumber", ctx, "+6281234567890").Return(&authentication.User{
		ID:          "u-123",
		Username:    "budi_santoso",
		PhoneNumber: "+6281234567890",
		Name:        "Budi Santoso",
	}, nil)

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AuthSession)
	assert.Equal(t, "acis_bot", resp.TelegramBotUsername)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_RequestOTP_SignIn_UnregisteredPhone(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegramclient.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_123", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		PhoneNumber: "081999888777",
		Action:      "login",
	}

	mockRepo.On("FindByPhoneNumber", ctx, "+6281999888777").Return(nil, nil)

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.ErrorIs(t, err, authentication.ErrPhoneNotRegistered)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_RequestOTP_SignUp_Success(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegramclient.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_123", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		PhoneNumber: "081299990000",
		Username:    "new_user",
		Action:      "register",
	}

	mockRepo.On("FindByPhoneNumber", ctx, "+6281299990000").Return(nil, nil)

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AuthSession)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_RequestOTP_SignUp_DuplicatePhone(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegramclient.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_123", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		PhoneNumber: "081234567890",
		Username:    "duplicate_guy",
		Action:      "register",
	}

	mockRepo.On("FindByPhoneNumber", ctx, "+6281234567890").Return(&authentication.User{
		ID:          "u-123",
		Username:    "already_registered",
		PhoneNumber: "+6281234567890",
	}, nil)

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.ErrorIs(t, err, authentication.ErrPhoneAlreadyRegistered)
	assert.Nil(t, resp)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_RequestOTP_TestUserBypass(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegramclient.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_12345678901234567890", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		PhoneNumber: "082123456781",
		Action:      "login",
	}

	mockRepo.On("FindByPhoneNumber", ctx, "+6282123456781").Return(&authentication.User{
		ID:          "test-admin-id",
		Username:    "admin_user",
		PhoneNumber: "+6282123456781",
		Name:        "Admin User",
	}, nil)

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.IsTestUser)
	assert.Equal(t, "123456", resp.TestOTP)

	// Verify Test Bypass
	verifyResp, err := authSvc.VerifyOTP(ctx, authentication.VerifyOTPReq{
		PhoneNumber: "082123456781",
		OTP:         "123456",
	})

	assert.NoError(t, err)
	assert.NotNil(t, verifyResp)
	assert.NotEmpty(t, verifyResp.Token)
	assert.Equal(t, "admin_user", verifyResp.User.Username)
	assert.Equal(t, "+6282123456781", verifyResp.User.PhoneNumber)
}

func TestAuthService_RequestOTP_InvalidPhoneFormat(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	otpCache := cache.NewOTPCache(rdb, "test_secret_key_32_bytes_long_123")
	tokenStore := cache.NewRefreshTokenStore(rdb)

	mockRepo := new(MockAuthRepo)
	tgClient := telegramclient.NewClient("")
	authSvc := authentication.NewService(mockRepo, nil, nil, otpCache, tokenStore, tgClient, nil, "jwt_secret_123", "acis_bot", 5*time.Minute)

	ctx := context.Background()
	req := authentication.RequestOTPReq{
		PhoneNumber: "12345", // Invalid format
	}

	resp, err := authSvc.RequestOTP(ctx, req)
	assert.ErrorIs(t, err, authentication.ErrInvalidPhoneFormat)
	assert.Nil(t, resp)
}
