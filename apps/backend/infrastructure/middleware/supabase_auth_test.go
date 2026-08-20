package middleware_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/middleware"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupabaseAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Generate ES256 P-256 keypair
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	pubJWK, err := jwk.FromRaw(privKey.PublicKey)
	require.NoError(t, err)
	_ = pubJWK.Set(jwk.KeyIDKey, "test-key-id")
	_ = pubJWK.Set(jwk.AlgorithmKey, jwa.ES256)
	_ = pubJWK.Set(jwk.KeyUsageKey, "sig")

	privJWK, err := jwk.FromRaw(privKey)
	require.NoError(t, err)
	_ = privJWK.Set(jwk.KeyIDKey, "test-key-id")
	_ = privJWK.Set(jwk.AlgorithmKey, jwa.ES256)

	keySet := jwk.NewSet()
	_ = keySet.AddKey(pubJWK)

	// 2. Mock JWKS server
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(keySet)
	}))
	defer jwksServer.Close()

	// 3. Initialize middleware
	ctx := context.Background()
	authMiddleware, err := middleware.NewSupabaseAuthMiddleware(ctx, jwksServer.URL)
	require.NoError(t, err)

	router := gin.New()
	router.Use(authMiddleware.Handler())
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("auth_user_id")
		email := c.GetString("auth_user_email")
		c.JSON(http.StatusOK, gin.H{"user_id": userID, "email": email})
	})

	t.Run("Valid Token Returns 200 and Populates Claims", func(t *testing.T) {
		builder := jwt.NewBuilder().
			Subject("user-uuid-123").
			Audience([]string{"authenticated"}).
			Expiration(time.Now().Add(1 * time.Hour)).
			IssuedAt(time.Now()).
			Claim("email", "test@example.com")

		token, err := builder.Build()
		require.NoError(t, err)

		signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, privJWK))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+string(signed))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "user-uuid-123", resp["user_id"])
		assert.Equal(t, "test@example.com", resp["email"])
	})

	t.Run("Expired Token Returns 401", func(t *testing.T) {
		builder := jwt.NewBuilder().
			Subject("user-uuid-123").
			Audience([]string{"authenticated"}).
			Expiration(time.Now().Add(-1 * time.Hour)).
			IssuedAt(time.Now().Add(-2 * time.Hour)).
			Claim("email", "test@example.com")

		token, err := builder.Build()
		require.NoError(t, err)

		signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, privJWK))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+string(signed))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Token Signed By Unknown Key Returns 401", func(t *testing.T) {
		otherKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)

		otherJWK, err := jwk.FromRaw(otherKey)
		require.NoError(t, err)
		_ = otherJWK.Set(jwk.KeyIDKey, "other-key-id")
		_ = otherJWK.Set(jwk.AlgorithmKey, jwa.ES256)

		builder := jwt.NewBuilder().
			Subject("user-uuid-123").
			Audience([]string{"authenticated"}).
			Expiration(time.Now().Add(1 * time.Hour)).
			IssuedAt(time.Now()).
			Claim("email", "test@example.com")

		token, err := builder.Build()
		require.NoError(t, err)

		signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, otherJWK))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+string(signed))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing Authorization Header Returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
