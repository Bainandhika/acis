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

	// 1. Bind and Validate JSON Body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Get User ID from Context (Nanti kita isi pas setup Auth middleware)
	// Untuk MVP, kita hardcode dulu user ID-nya
	createdBy := "system-admin-temp"

	// 3. Call Service
	wallet, err := h.walletService.CreateWallet(c.Request.Context(), req, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Return Success Response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Wallet created successfully",
		"data":    wallet,
	})
}
