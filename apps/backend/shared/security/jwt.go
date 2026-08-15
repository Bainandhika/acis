package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims extends jwt.RegisteredClaims with custom fields
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"` // 'admin' or 'member'
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for the user with given duration
func GenerateToken(userID, role, secret string, duration time.Duration) (string, error) {
	if duration <= 0 {
		duration = 15 * time.Minute
	}
	expirationTime := time.Now().Add(duration)
	claims := &CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "acis-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateAccessToken generates a standard 15-minute access token
func GenerateAccessToken(userID, role, secret string) (string, error) {
	return GenerateToken(userID, role, secret, 15*time.Minute)
}

// ValidateToken parses and validates the JWT token string using the provided secret
func ValidateToken(tokenString, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
