package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Bainandhika/acis/apps/backend/internal/family"
	"github.com/Bainandhika/acis/apps/backend/internal/transaction"
	zerolog "github.com/rs/zerolog/log"
)

type BotService struct {
	txService     transaction.Service
	familyService family.Service
	client        *Client
}

func NewBotService(txService transaction.Service, familyService family.Service, client *Client) *BotService {
	return &BotService{
		txService:     txService,
		familyService: familyService,
		client:        client,
	}
}

func (s *BotService) ProcessUpdate(ctx context.Context, update *TelegramUpdate) (string, error) {
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
		return s.handleCatat(ctx, update, parts)
	case "/saldo":
		return s.handleSaldo(ctx, update)
	default:
		return "Perintah tidak dikenali. Gunakan `/catat [dompet] [nominal] [keterangan]` atau `/saldo`.", nil
	}
}

func (s *BotService) handleCatat(ctx context.Context, update *TelegramUpdate, parts []string) (string, error) {
	// Format: /catat [wallet_id_or_name] [amount] [notes...]
	if len(parts) < 4 {
		return "Format salah. Gunakan: `/catat [wallet_id] [nominal] [keterangan]`", nil
	}

	walletID := parts[1]
	amountStr := parts[2]
	description := strings.Join(parts[3:], " ")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return "Nominal transaksi tidak valid.", nil
	}

	req := transaction.CreateTransactionDTO{
		WalletID:    walletID,
		UserID:      fmt.Sprintf("telegram:%d", update.Message.Chat.ID),
		Type:        "expense",
		Amount:      amount,
		Category:    "telegram_catat",
		Description: &description,
	}

	res, err := s.txService.CreateDirectTransaction(ctx, req)
	if err != nil {
		zerolog.Error().Err(err).Msg("Failed to process /catat via Telegram")
		return fmt.Sprintf("❌ Gagal mencatat transaksi: %s", err.Error()), nil
	}

	return fmt.Sprintf("✅ Transaksi tersimpan! ID: `%s`, Nominal: Rp %.2f", res.ID, res.Amount), nil
}

func (s *BotService) handleSaldo(ctx context.Context, update *TelegramUpdate) (string, error) {
	// Query balances (using default or mapped family ID)
	familyID := "default"
	balances, err := s.familyService.GetWalletBalances(ctx, familyID)
	if err != nil {
		return "⚠️ Gagal mengambil saldo dompet.", nil
	}

	if len(balances) == 0 {
		return "Belum ada dompet terdaftar.", nil
	}

	var sb strings.Builder
	sb.WriteString("📊 *Ringkasan Saldo Dompet Keluarga*\n\n")
	for _, b := range balances {
		sb.WriteString(fmt.Sprintf("• *%s*: Rp %.2f (Min: Rp %.2f)\n", b.WalletName, b.CurrentBalance, b.MinimumLimit))
	}

	return sb.String(), nil
}
