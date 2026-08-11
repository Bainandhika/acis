package cache_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/Bainandhika/acis/apps/backend/internal/shared/cache"
)

func TestOTPCache_VerifySuccess(t *testing.T) {
	otpCache := cache.NewOTPCache(1 * time.Minute)
	defer otpCache.Close()

	email := "user@example.com"
	code := "123456"

	otpCache.StoreOTP(email, code, 1*time.Minute)

	ok, err := otpCache.VerifyOTP(email, code)
	assert.NoError(t, err)
	assert.True(t, ok)

	// Verify consumed
	_, err = otpCache.VerifyOTP(email, code)
	assert.ErrorIs(t, err, cache.ErrOTPNotFound)
}

func TestOTPCache_LockoutAfter3Failures(t *testing.T) {
	otpCache := cache.NewOTPCache(1 * time.Minute)
	defer otpCache.Close()

	email := "lockout@example.com"
	correctCode := "654321"
	wrongCode := "000000"

	otpCache.StoreOTP(email, correctCode, 1*time.Minute)

	// 1st attempt failure
	_, err := otpCache.VerifyOTP(email, wrongCode)
	assert.ErrorIs(t, err, cache.ErrInvalidOTP)

	// 2nd attempt failure
	_, err = otpCache.VerifyOTP(email, wrongCode)
	assert.ErrorIs(t, err, cache.ErrInvalidOTP)

	// 3rd attempt failure -> lockout
	_, err = otpCache.VerifyOTP(email, wrongCode)
	assert.ErrorIs(t, err, cache.ErrOTPLockedOut)

	// Subsequent attempt (even with correct code) fails due to lockout
	_, err = otpCache.VerifyOTP(email, correctCode)
	assert.ErrorIs(t, err, cache.ErrOTPLockedOut)
}
