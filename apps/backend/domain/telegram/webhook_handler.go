package telegram

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	botService *BotService
	client     *Client
}

func NewWebhookHandler(botService *BotService, client *Client) *WebhookHandler {
	return &WebhookHandler{
		botService: botService,
		client:     client,
	}
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	expectedSecret := os.Getenv("TELEGRAM_WEBHOOK_SECRET")
	if expectedSecret != "" {
		headerSecret := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
		if headerSecret != expectedSecret {
			slog.Warn("Unauthorized Telegram webhook request: secret token mismatch")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized webhook token"})
			c.Abort()
			return
		}
	}

	var update TelegramUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	replyMsg, err := h.botService.ProcessUpdate(c.Request.Context(), &update)
	if err != nil {
		slog.Error("Error processing Telegram update", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "processing error"})
		return
	}

	if replyMsg != "" && update.Message != nil {
		_ = h.client.SendMessage(c.Request.Context(), update.Message.Chat.ID, replyMsg)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
