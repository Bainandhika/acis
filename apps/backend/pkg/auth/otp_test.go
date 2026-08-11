package auth

import (
	"strconv"
	"testing"
)

func TestGenerateOTP(t *testing.T) {
	otp, err := GenerateOTP()
	if err != nil {
		t.Fatalf("GenerateOTP failed: %v", err)
	}

	if len(otp) != OTPLength {
		t.Errorf("Expected OTP length %d, got %d", OTPLength, len(otp))
	}

	num, err := strconv.Atoi(otp)
	if err != nil {
		t.Fatalf("OTP is not a valid number: %v", err)
	}

	if num < 100000 || num > 999999 {
		t.Errorf("OTP out of range [100000, 999999]: %d", num)
	}
}

func TestHashAndVerifyOTP(t *testing.T) {
	otp := "123456"
	hash, err := HashOTP(otp)
	if err != nil {
		t.Fatalf("HashOTP failed: %v", err)
	}

	if !VerifyOTP(otp, hash) {
		t.Error("VerifyOTP failed for matching OTP and hash")
	}

	if VerifyOTP("654321", hash) {
		t.Error("VerifyOTP succeeded for non-matching OTP")
	}
}
