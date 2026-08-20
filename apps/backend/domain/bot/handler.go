package bot

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Bainandhika/acis/apps/backend/domain/authentication"
	"github.com/Bainandhika/acis/apps/backend/domain/family"
	"github.com/Bainandhika/acis/apps/backend/domain/transaction"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/gin-gonic/gin"
)

type BotHandler struct {
	familySvc family.FamilyService
	txSvc     transaction.TransactionService
	authRepo  authentication.AuthRepository
	linkStore *cache.TelegramLinkStore
}

func NewBotHandler(familySvc family.FamilyService, txSvc transaction.TransactionService, authRepo authentication.AuthRepository, linkStore *cache.TelegramLinkStore) *BotHandler {
	return &BotHandler{
		familySvc: familySvc,
		txSvc:     txSvc,
		authRepo:  authRepo,
		linkStore: linkStore,
	}
}

// GenerateLinkCode generates a 6-character code for the user to link their Telegram account
func (h *BotHandler) GenerateLinkCode(c *gin.Context) {
	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	code, err := h.linkStore.GenerateCode(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate linking code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":       code,
		"expires_in": 600,
	})
}

type InternalLinkReq struct {
	Code   string `json:"code" binding:"required"`
	ChatID int64  `json:"chat_id" binding:"required"`
}

// LinkAccount consumes the link code and binds the user's telegram_chat_id
func (h *BotHandler) LinkAccount(c *gin.Context) {
	var req InternalLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.linkStore.PopCode(c.Request.Context(), strings.ToUpper(strings.TrimSpace(req.Code)))
	if err != nil {
		if errors.Is(err, cache.ErrLinkCodeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "link code not found or expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process link code"})
		return
	}

	if err := h.authRepo.UpdateTelegramChatID(c.Request.Context(), userID, req.ChatID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link telegram account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account linked successfully",
		"user_id": userID,
		"chat_id": req.ChatID,
	})
}

type LinkReq struct {
	InviteCode string `json:"invite_code" binding:"required"`
	ChatID     int64  `json:"chat_id" binding:"required"`
}

func (h *BotHandler) Link(c *gin.Context) {
	var req LinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.familySvc.LinkTelegramChatID(c.Request.Context(), strings.ToUpper(req.InviteCode), req.ChatID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Family linked successfully to Telegram chat"})
}

func (h *BotHandler) GetFamily(c *gin.Context) {
	chatIDStr := c.Query("chat_id")
	if chatIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id is required"})
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	fam, err := h.familySvc.FindByTelegramChatID(c.Request.Context(), chatID)
	if err != nil || fam == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "family not found for this telegram chat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fam})
}

func (h *BotHandler) Balance(c *gin.Context) {
	chatIDStr := c.Query("chat_id")
	if chatIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id is required"})
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	fam, err := h.familySvc.FindByTelegramChatID(c.Request.Context(), chatID)
	if err != nil || fam == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "family not found for this telegram chat"})
		return
	}

	balances, err := h.familySvc.GetWalletBalances(c.Request.Context(), fam.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve wallet balances"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": balances})
}

type RecordTxReq struct {
	WalletShortID string  `json:"wallet_short_id" binding:"required"`
	Type          string  `json:"type" binding:"required,oneof=income expense"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Description   string  `json:"description"`
	ChatID        int64   `json:"chat_id" binding:"required"`
}

func (h *BotHandler) RecordTransaction(c *gin.Context) {
	var req RecordTxReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fam, err := h.familySvc.FindByTelegramChatID(c.Request.Context(), req.ChatID)
	if err != nil || fam == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chat is not linked to any family"})
		return
	}

	wallet, err := h.familySvc.FindWalletByShortID(c.Request.Context(), strings.ToUpper(req.WalletShortID))
	if err != nil || wallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("wallet with ID '%s' not found", req.WalletShortID)})
		return
	}

	if wallet.FamilyID != fam.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "wallet does not belong to your family"})
		return
	}

	members, err := h.familySvc.GetMembers(c.Request.Context(), fam.ID)
	if err != nil || len(members) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no family members found"})
		return
	}

	adminUserID := members[0].UserID
	for _, m := range members {
		if m.Role == "admin" {
			adminUserID = m.UserID
			break
		}
	}

	desc := req.Description
	res, err := h.txSvc.CreateDirectTransaction(c.Request.Context(), transaction.CreateTransactionDTO{
		WalletID:    wallet.ID,
		UserID:      adminUserID,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: &desc,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Transaction recorded successfully",
		"data":    res,
	})
}
