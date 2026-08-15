package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

type RefreshSession struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type RefreshTokenStore struct {
	client *redis.Client
}

func NewRefreshTokenStore(client *redis.Client) *RefreshTokenStore {
	return &RefreshTokenStore{client: client}
}

// GenerateSecureToken creates a cryptographically secure random token string
func GenerateSecureToken(byteLen int) (string, error) {
	if byteLen <= 0 {
		byteLen = 32
	}
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (s *RefreshTokenStore) formatKey(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}

func (s *RefreshTokenStore) formatUserSetKey(userID string) string {
	return fmt.Sprintf("user_refresh_tokens:%s", userID)
}

// StoreRefreshToken stores a new refresh token with TTL and tracks it under the user
func (s *RefreshTokenStore) StoreRefreshToken(ctx context.Context, token string, session RefreshSession, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	session.CreatedAt = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh session: %w", err)
	}

	key := s.formatKey(token)
	pipe := s.client.Pipeline()
	pipe.Set(ctx, key, data, ttl)
	userKey := s.formatUserSetKey(session.UserID)
	pipe.SAdd(ctx, userKey, token)
	pipe.Expire(ctx, userKey, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to store refresh token in redis: %w", err)
	}
	return nil
}

// GetAndRevokeRefreshToken retrieves session data and atomically deletes the token (single-use rotation)
func (s *RefreshTokenStore) GetAndRevokeRefreshToken(ctx context.Context, token string) (*RefreshSession, error) {
	key := s.formatKey(token)

	// Fetch data first
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("failed to query redis for refresh token: %w", err)
	}

	var session RefreshSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to decode refresh session: %w", err)
	}

	// Single-use: delete old token and remove from user set immediately
	pipe := s.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, s.formatUserSetKey(session.UserID), token)
	_, _ = pipe.Exec(ctx)

	return &session, nil
}

// RevokeRefreshToken revokes a single refresh token (e.g. on logout)
func (s *RefreshTokenStore) RevokeRefreshToken(ctx context.Context, token string) error {
	key := s.formatKey(token)

	data, err := s.client.Get(ctx, key).Bytes()
	if err == nil {
		var session RefreshSession
		if json.Unmarshal(data, &session) == nil {
			s.client.SRem(ctx, s.formatUserSetKey(session.UserID), token)
		}
	}

	return s.client.Del(ctx, key).Err()
}

// RevokeAllUserTokens invalidates all active refresh tokens for a user (instant complete logout)
func (s *RefreshTokenStore) RevokeAllUserTokens(ctx context.Context, userID string) error {
	userKey := s.formatUserSetKey(userID)
	tokens, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to fetch user tokens: %w", err)
	}

	if len(tokens) > 0 {
		keys := make([]string, 0, len(tokens)+1)
		for _, tok := range tokens {
			keys = append(keys, s.formatKey(tok))
		}
		keys = append(keys, userKey)
		return s.client.Del(ctx, keys...).Err()
	}

	return s.client.Del(ctx, userKey).Err()
}
