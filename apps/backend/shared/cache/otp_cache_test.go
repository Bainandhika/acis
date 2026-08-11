package cache

import (
	"errors"
	"testing"
	"time"
)

func TestOTPCache_StoreAndVerify(t *testing.T) {
	c := NewOTPCache(1 * time.Minute)
	defer c.Close()

	email := "test@example.com"
	code := "123456"

	c.StoreOTP(email, code, 1*time.Minute)

	// Verify wrong code
	ok, err := c.VerifyOTP(email, "000000")
	if ok || !errors.Is(err, ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP, got ok=%v, err=%v", ok, err)
	}

	// Verify correct code
	ok, err = c.VerifyOTP(email, code)
	if !ok || err != nil {
		t.Fatalf("expected success, got ok=%v, err=%v", ok, err)
	}
}
