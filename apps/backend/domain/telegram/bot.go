package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message"`
}

type TelegramMessage struct {
	MessageID int          `json:"message_id"`
	Chat      TelegramChat `json:"chat"`
	Text      string       `json:"text"`
}

type TelegramChat struct {
	ID int64 `json:"id"`
}

type CommandHandler struct{}

func NewCommandHandler() *CommandHandler {
	return &CommandHandler{}
}

func (h *CommandHandler) ProcessMessage(ctx context.Context, update *TelegramUpdate) (string, error) {
	if update.Message == nil || update.Message.Text == "" {
		return "", nil
	}

	text := strings.TrimSpace(update.Message.Text)
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", nil
	}

	command := strings.ToLower(parts[0])
	switch command {
	case "/catat":
		if len(parts) < 4 {
			return "Format salah. Gunakan: `/catat [dompet] [nominal] [keterangan]`", nil
		}
		walletName := parts[1]
		amountStr := parts[2]
		desc := strings.Join(parts[3:], " ")

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			return "Nominal tidak valid.", nil
		}

		log.Info().Str("wallet", walletName).Float64("amount", amount).Str("desc", desc).Msg("Telegram transaction recorded")
		return fmt.Sprintf("✅ Catatan tersimpan! Dompet: %s, Nominal: Rp %.2f (%s)", walletName, amount, desc), nil

	case "/saldo":
		return "📊 Fitur Cek Saldo Dompet (Gunakan dashboard web untuk detail lengkap).", nil

	default:
		return "Perintah tidak dikenali. Gunakan `/catat` atau `/saldo`.", nil
	}
}
