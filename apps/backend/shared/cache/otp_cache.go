package cache

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrOTPNotFound  = errors.New("otp not found or expired")
	ErrOTPLockedOut = errors.New("too many invalid otp attempts, locked out")
	ErrInvalidOTP   = errors.New("invalid otp code")
)

type OTPEntry struct {
	Code         string
	AttemptsLeft int
	LockoutUntil time.Time
}

type OTPCache struct {
	cache *TTLCache[string, OTPEntry]
}

func NewOTPCache(cleanupInterval time.Duration) *OTPCache {
	return &OTPCache{
		cache: NewTTLCache[string, OTPEntry](cleanupInterval),
	}
}

func (o *OTPCache) StoreOTP(email, code string, ttl time.Duration) {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	entry := OTPEntry{
		Code:         code,
		AttemptsLeft: 3,
	}
	o.cache.Set(email, entry, ttl)
}

func (o *OTPCache) VerifyOTP(email, code string) (bool, error) {
	entry, exists := o.cache.Get(email)
	if !exists {
		return false, ErrOTPNotFound
	}

	if time.Now().Before(entry.LockoutUntil) {
		return false, ErrOTPLockedOut
	}

	if entry.Code == code {
		o.cache.Delete(email)
		return true, nil
	}

	entry.AttemptsLeft--
	if entry.AttemptsLeft <= 0 {
		entry.LockoutUntil = time.Now().Add(5 * time.Minute)
		o.cache.Set(email, entry, 5*time.Minute)
		return false, ErrOTPLockedOut
	}

	o.cache.Set(email, entry, 5*time.Minute)
	return false, fmt.Errorf("%w: %d attempts remaining", ErrInvalidOTP, entry.AttemptsLeft)
}

func (o *OTPCache) Delete(email string) {
	o.cache.Delete(email)
}

func (o *OTPCache) Close() {
	o.cache.Close()
}
