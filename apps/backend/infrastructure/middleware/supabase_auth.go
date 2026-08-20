package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type SupabaseAuthMiddleware struct {
	cache   *jwk.Cache
	jwksURL string
}

func NewSupabaseAuthMiddleware(ctx context.Context, jwksURL string) (*SupabaseAuthMiddleware, error) {
	c := jwk.NewCache(ctx)
	if err := c.Register(jwksURL, jwk.WithMinRefreshInterval(15*time.Minute)); err != nil {
		return nil, fmt.Errorf("failed to register JWKS URL %s: %w", jwksURL, err)
	}

	// Refresh once at boot; fail fast on error
	if _, err := c.Refresh(ctx, jwksURL); err != nil {
		return nil, fmt.Errorf("failed to initial refresh JWKS from %s: %w", jwksURL, err)
	}

	return &SupabaseAuthMiddleware{
		cache:   c,
		jwksURL: jwksURL,
	}, nil
}

func (m *SupabaseAuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		rawToken := strings.TrimPrefix(authHeader, "Bearer ")
		if rawToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		set, err := m.cache.Get(ctx, m.jwksURL)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to retrieve public keys"})
			c.Abort()
			return
		}

		// JWKS-backed verification: validates ES256 (ECC P-256) signatures and
		// survives Supabase key rotation via cached refresh. Never store a shared
		// JWT signing secret in this service (asymmetric trust model, OWASP A02).
		tok, err := jwt.Parse([]byte(rawToken),
			jwt.WithKeySet(set),
			jwt.WithValidate(true),
			jwt.WithAcceptableSkew(5*time.Second),
			jwt.WithAudience("authenticated"),
		)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		userID := tok.Subject()
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing subject claim"})
			c.Abort()
			return
		}

		var email string
		if emailClaim, ok := tok.Get("email"); ok {
			if emailStr, ok := emailClaim.(string); ok {
				email = emailStr
			}
		}

		// Inject into standard context.Context for the database layer to extract
		reqCtx := c.Request.Context()
		reqCtx = context.WithValue(reqCtx, "auth_user_id", userID)
		reqCtx = context.WithValue(reqCtx, "auth_user_email", email)
		c.Request = c.Request.WithContext(reqCtx)

		c.Set("auth_user_id", userID)
		c.Set("auth_user_email", email)

		// Set legacy keys as fallback compatibility during migration
		c.Set("user_id", userID)
		c.Set("user_email", email)

		c.Next()
	}
}
