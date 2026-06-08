package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	OTPLength = 6
	OTPExpiry = 5 * time.Minute
)

// GenerateOTP creates a cryptographically secure 6-digit OTP
func GenerateOTP() (string, error) {
	// Generate a random number between 100000 and 999999
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	
	otp := n.Int64() + 100000
	return fmt.Sprintf("%d", otp), nil
}

// HashOTP hashes the OTP using bcrypt before storing it in the DB
// NEVER store plain text OTP in the database (OWASP A02)
func HashOTP(otp string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyOTP compares a plain text OTP with its bcrypt hash
func VerifyOTP(otp, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(otp))
	return err == nil
}