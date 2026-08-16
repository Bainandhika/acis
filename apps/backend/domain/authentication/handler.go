package authentication

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	RefreshTokenCookieName = "refresh_token"
	AuthCookiePath         = "/api/v1/authentication"
	RefreshTokenTTLSeconds = 7 * 24 * 3600 // 7 days
)

type AuthHandler struct {
	authService  AuthService
	isProduction bool
}

func NewAuthHandler(authService AuthService, isProduction bool) *AuthHandler {
	return &AuthHandler{authService: authService, isProduction: isProduction}
}

func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req RequestOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email dan nomor telepon yang valid wajib diisi"})
		return
	}

	resp, err := h.authService.RequestOTP(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailMismatch), errors.Is(err, ErrPhoneMismatch), errors.Is(err, ErrAccountConflict):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, ErrTooManyRequests):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email, nomor telepon, dan 6-digit OTP wajib diisi"})
		return
	}

	resp, err := h.authService.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailMismatch), errors.Is(err, ErrPhoneMismatch):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}
		return
	}

	// Set rotating refresh token in HttpOnly, Secure, SameSite, path-scoped cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		RefreshTokenCookieName,
		resp.RefreshToken,
		RefreshTokenTTLSeconds,
		AuthCookiePath,
		"",
		h.isProduction,
		true, // HttpOnly
	)

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(RefreshTokenCookieName)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token cookie"})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		// Clear invalid cookie on failure
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(RefreshTokenCookieName, "", -1, AuthCookiePath, "", h.isProduction, true)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set new rotated refresh token in cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		RefreshTokenCookieName,
		resp.RefreshToken,
		RefreshTokenTTLSeconds,
		AuthCookiePath,
		"",
		h.isProduction,
		true, // HttpOnly
	)

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(RefreshTokenCookieName)
	if refreshToken != "" {
		_ = h.authService.Logout(c.Request.Context(), refreshToken)
	}

	// Clear HttpOnly cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(RefreshTokenCookieName, "", -1, AuthCookiePath, "", h.isProduction, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
