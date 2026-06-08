package handler

import (
	"net/http"

	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RequestOTP handles POST /api/v1/auth/request-otp
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req dto.RequestOTPRequest
	// Bind JSON and validate (email format)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Service
	if err := h.authService.RequestOTP(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully. Check your email (or logs for MVP).",
	})
}

// VerifyOTP handles POST /api/v1/auth/verify-otp
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call Service to verify and get JWT
	resp, err := h.authService.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
