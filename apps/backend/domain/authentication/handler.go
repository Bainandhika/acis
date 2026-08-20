package authentication

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Provision provisions user profile on initial sign in
func (h *AuthHandler) Provision(c *gin.Context) {
	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	email := c.GetString("auth_user_email")

	var req ProvisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Tolerant body binding (e.g. empty body uses default values)
		req = ProvisionRequest{}
	}

	profile, err := h.authService.Provision(c.Request.Context(), userID, email, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// Me returns the current authenticated user's profile and family memberships
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	email := c.GetString("auth_user_email")

	profile, err := h.authService.GetMe(c.Request.Context(), userID, email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}
