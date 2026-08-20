package cache

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrLinkCodeNotFound = errors.New("link code not found or expired")

type TelegramLinkStore struct {
	client *redis.Client
}

func NewTelegramLinkStore(client *redis.Client) *TelegramLinkStore {
	return &TelegramLinkStore{client: client}
}

// GenerateCode creates a random 6-character alphanumeric uppercase code and stores it in Redis for 10 minutes (600s).
func (s *TelegramLinkStore) GenerateCode(ctx context.Context, userID string) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // omit ambiguous 0/O, 1/I
	codeBytes := make([]byte, 6)
	for i := range codeBytes {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random link code: %w", err)
		}
		codeBytes[i] = charset[num.Int64()]
	}
	code := string(codeBytes)

	key := fmt.Sprintf("link:%s", code)
	if err := s.client.Set(ctx, key, userID, 10*time.Minute).Err(); err != nil {
		return "", fmt.Errorf("failed to save link code: %w", err)
	}

	return code, nil
}

// PopCode retrieves and deletes the userID associated with the code.
func (s *TelegramLinkStore) PopCode(ctx context.Context, code string) (string, error) {
	key := fmt.Sprintf("link:%s", code)
	val, err := s.client.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) || val == "" {
		return "", ErrLinkCodeNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to retrieve link code: %w", err)
	}
	return val, nil
}
