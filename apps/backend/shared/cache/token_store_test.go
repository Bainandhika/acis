package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	return client, s
}

func TestRefreshTokenStore_StoreAndRotate(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	store := NewRefreshTokenStore(client)

	token, err := GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	session := RefreshSession{
		UserID: "user-123",
		Role:   "admin",
		Email:  "user@example.com",
	}

	// 1. Store refresh token
	err = store.StoreRefreshToken(ctx, token, session, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("failed to store refresh token: %v", err)
	}

	// 2. Rotate (single-use retrieve and revoke)
	retrieved, err := store.GetAndRevokeRefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("failed to get and revoke refresh token: %v", err)
	}
	if retrieved.UserID != session.UserID {
		t.Errorf("expected user id %s, got %s", session.UserID, retrieved.UserID)
	}
	if retrieved.Role != session.Role {
		t.Errorf("expected role %s, got %s", session.Role, retrieved.Role)
	}

	// 3. Second rotation attempt with old token must fail (single-use!)
	_, err = store.GetAndRevokeRefreshToken(ctx, token)
	if err == nil {
		t.Errorf("expected error when reusing rotated token, got nil")
	}
}

func TestRefreshTokenStore_RevokeToken(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	store := NewRefreshTokenStore(client)

	token, _ := GenerateSecureToken(32)
	session := RefreshSession{
		UserID: "user-456",
		Role:   "member",
		Email:  "member@example.com",
	}

	_ = store.StoreRefreshToken(ctx, token, session, 1*time.Hour)

	// Revoke token
	err := store.RevokeRefreshToken(ctx, token)
	if err != nil {
		t.Fatalf("failed to revoke refresh token: %v", err)
	}

	// Verify it cannot be retrieved
	_, err = store.GetAndRevokeRefreshToken(ctx, token)
	if err == nil {
		t.Errorf("expected error retrieving revoked token, got nil")
	}
}
