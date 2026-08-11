package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "super-secret-key-12345"
	userID := "user-123"
	role := "admin"

	// 1. Generate Token
	tokenString, err := GenerateToken(userID, role, secret, 1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if tokenString == "" {
		t.Fatal("Expected non-empty token string")
	}

	// 2. Validate Token with correct secret
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}

	// 3. Validate Token with wrong secret
	_, err = ValidateToken(tokenString, "wrong-secret-key")
	if err == nil {
		t.Error("Expected error when validating token with wrong secret, got nil")
	}
}

func TestValidateExpiredToken(t *testing.T) {
	secret := "secret-key"
	claims := &CustomClaims{
		UserID: "user-abc",
		Role:   "member",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "acis-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	_, err = ValidateToken(tokenString, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}
