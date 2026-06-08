package handler

import (
	"net/http"

	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletService service.WalletService
}

// NewWalletHandler creates a new WalletHandler
func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

// CreateWallet handles POST /api/v1/wallets
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req dto.CreateWalletRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get User ID from Context (Injected by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Call Service with real User ID
	wallet, err := h.walletService.CreateWallet(c.Request.Context(), req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Wallet created successfully",
		"data":    wallet,
	})
}
