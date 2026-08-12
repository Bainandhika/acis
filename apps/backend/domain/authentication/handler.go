package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.RequestOTP(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully. Check your email.",
	})
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("auth_token", resp.Token, 86400, "/", "", h.isProduction, true)

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("auth_token", "", -1, "/", "", h.isProduction, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

