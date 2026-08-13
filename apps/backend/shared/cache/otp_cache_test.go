package cache

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisOTPCache_StoreVerifyAndEncryption(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	secretKey := "test_secret_key_32_bytes_long_123"
	cache := NewOTPCache(rdb, secretKey)

	ctx := context.Background()
	email := "user@example.com"
	otp := "654321"

	// 1. Store OTP
	err := cache.StoreOTP(ctx, email, otp, 5*time.Minute)
	if err != nil {
		t.Fatalf("StoreOTP failed: %v", err)
	}

	// 2. Verify raw data in Redis is ONLY encrypted (not plain OTP)
	raw, err := rdb.Get(ctx, "otp:"+email).Result()
	if err != nil {
		t.Fatalf("redis get failed: %v", err)
	}

	var entry OTPEntry
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if entry.EncryptedCode == otp {
		t.Fatal("SECURITY VIOLATION: OTP is stored in plaintext in Redis!")
	}

	// 3. Verify wrong OTP code
	ok, err := cache.VerifyOTP(ctx, email, "000000")
	if ok || !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP, got ok=%v, err=%v", ok, err)
	}

	// 4. Verify correct OTP code
	ok, err = cache.VerifyOTP(ctx, email, otp)
	if !ok || err != nil {
		t.Fatalf("expected verification success, got ok=%v, err=%v", ok, err)
	}

	// 5. Verify key is deleted after successful verification
	exists := rdb.Exists(ctx, "otp:"+email).Val()
	if exists != 0 {
		t.Fatal("expected key to be deleted after successful verification")
	}
}

func TestRedisOTPCache_RateLimiting(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cache := NewOTPCache(rdb, "secret_key_32_bytes_long_string!!")
	ctx := context.Background()
	email := "limit@example.com"

	for i := 0; i < 3; i++ {
		can, err := cache.CanRequestOTP(ctx, email)
		if err != nil || !can {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	// 4th attempt should be blocked
	can, err := cache.CanRequestOTP(ctx, email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if can {
		t.Fatal("expected 4th request to be rate limited")
	}
}
