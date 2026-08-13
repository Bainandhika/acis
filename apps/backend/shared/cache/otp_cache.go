package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Bainandhika/acis/apps/backend/shared/security"
	"github.com/redis/go-redis/v9"
)

var (
	ErrOTPNotFound  = errors.New("otp not found or expired")
	ErrOTPLockedOut = errors.New("too many invalid otp attempts, locked out")
	ErrInvalidOTP   = errors.New("invalid otp code")
)

type OTPEntry struct {
	EncryptedCode string    `json:"encrypted_code"`
	AttemptsLeft  int       `json:"attempts_left"`
	LockoutUntil  time.Time `json:"lockout_until"`
}

type OTPCache struct {
	client    redis.Cmdable
	secretKey string
}

func NewOTPCache(client redis.Cmdable, secretKey string) *OTPCache {
	return &OTPCache{
		client:    client,
		secretKey: secretKey,
	}
}

func (o *OTPCache) getKey(email string) string {
	return "otp:" + email
}

func (o *OTPCache) getRateKey(email string) string {
	return "otp:rate:" + email
}

// StoreOTP encrypts the OTP using AES-GCM and stores only the encrypted payload in standalone Redis with configurable TTL
func (o *OTPCache) StoreOTP(ctx context.Context, email, plainCode string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	encrypted, err := security.EncryptAESGCM(plainCode, o.secretKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt OTP: %w", err)
	}

	entry := OTPEntry{
		EncryptedCode: encrypted,
		AttemptsLeft:  3,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return o.client.Set(ctx, o.getKey(email), data, ttl).Err()
}

// CanRequestOTP checks rate limits using standalone Redis
func (o *OTPCache) CanRequestOTP(ctx context.Context, email string) (bool, error) {
	key := o.getRateKey(email)
	count, err := o.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		o.client.Expire(ctx, key, 15*time.Minute)
	}
	if count > 3 {
		return false, nil
	}
	return true, nil
}

// VerifyOTP fetches the encrypted OTP payload from Redis, decrypts it, and verifies match
func (o *OTPCache) VerifyOTP(ctx context.Context, email, plainCode string) (bool, error) {
	key := o.getKey(email)
	val, err := o.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, ErrOTPNotFound
	} else if err != nil {
		return false, err
	}

	var entry OTPEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		return false, err
	}

	if !entry.LockoutUntil.IsZero() && time.Now().Before(entry.LockoutUntil) {
		return false, ErrOTPLockedOut
	}

	decryptedCode, err := security.DecryptAESGCM(entry.EncryptedCode, o.secretKey)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt OTP: %w", err)
	}

	if decryptedCode == plainCode {
		o.client.Del(ctx, key)
		return true, nil
	}

	entry.AttemptsLeft--
	if entry.AttemptsLeft <= 0 {
		entry.LockoutUntil = time.Now().Add(5 * time.Minute)
		data, _ := json.Marshal(entry)
		o.client.Set(ctx, key, data, 5*time.Minute)
		return false, ErrOTPLockedOut
	}

	data, _ := json.Marshal(entry)
	ttl, _ := o.client.TTL(ctx, key).Result()
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	o.client.Set(ctx, key, data, ttl)
	return false, fmt.Errorf("%w: %d attempts remaining", ErrInvalidOTP, entry.AttemptsLeft)
}

func (o *OTPCache) Delete(ctx context.Context, email string) error {
	return o.client.Del(ctx, o.getKey(email)).Err()
}
