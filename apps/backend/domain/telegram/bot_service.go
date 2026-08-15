package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

type BotService struct {
	txService     TransactionService
	familyService FamilyService
	client        *Client
}

func NewBotService(txService TransactionService, familyService FamilyService, client *Client) *BotService {
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
	case "/link":
		return s.handleLink(ctx, update, parts)
	default:
		return "Perintah tidak dikenali. Gunakan `/catat [wallet_id] [nominal] [keterangan]`, `/saldo`, atau `/link [invite_code]`.", nil
	}
}

func (s *BotService) handleLink(ctx context.Context, update *TelegramUpdate, parts []string) (string, error) {
	if len(parts) < 2 {
		return "Format salah. Gunakan: `/link [invite_code]`", nil
	}
	inviteCode := strings.ToUpper(parts[1])
	chatID := update.Message.Chat.ID
	
	if err := s.familyService.LinkTelegramChatID(ctx, inviteCode, chatID); err != nil {
		slog.Error("Failed to link Telegram chat", slog.Any("error", err), slog.Int64("chat_id", chatID), slog.String("invite_code", inviteCode))
		return "❌ Gagal menghubungkan Telegram. Pastikan Kode Invite benar.", nil
	}

	slog.Info("Telegram link command succeeded", slog.Int64("chat_id", chatID), slog.String("invite_code", inviteCode))
	return fmt.Sprintf("✅ Chat Telegram ini berhasil dihubungkan dengan keluarga untuk kode invite `%s`!", inviteCode), nil
}

func (s *BotService) handleCatat(ctx context.Context, update *TelegramUpdate, parts []string) (string, error) {
	if len(parts) < 4 {
		return "Format salah. Gunakan: `/catat [wallet_id] [nominal] [keterangan]`", nil
	}

	chatID := update.Message.Chat.ID
	fam, err := s.familyService.FindByTelegramChatID(ctx, chatID)
	if err != nil || fam == nil {
		return "⚠️ Chat Telegram ini belum terhubung dengan keluarga. Gunakan `/link [invite_code]` terlebih dahulu.", nil
	}

	members, err := s.familyService.GetMembers(ctx, fam.ID)
	if err != nil || len(members) == 0 {
		return "⚠️ Tidak dapat menemukan anggota keluarga.", nil
	}

	adminUserID := members[0].UserID
	for _, m := range members {
		if m.Role == "admin" {
			adminUserID = m.UserID
			break
		}
	}

	walletID := parts[1]
	amountStr := parts[2]
	description := strings.Join(parts[3:], " ")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return "Nominal transaksi tidak valid.", nil
	}

	balances, err := s.familyService.GetWalletBalances(ctx, fam.ID)
	if err != nil {
		return "⚠️ Gagal memverifikasi dompet.", nil
	}
	walletValid := false
	for _, b := range balances {
		if b.WalletID == walletID {
			walletValid = true
			break
		}
	}
	if !walletValid {
		return "❌ Wallet ID tidak ditemukan dalam keluarga Anda. Gunakan `/saldo` untuk melihat daftar Wallet ID.", nil
	}

	req := CreateTransactionDTO{
		WalletID:    walletID,
		UserID:      adminUserID,
		Type:        "expense",
		Amount:      amount,
		Category:    "telegram_catat",
		Description: &description,
	}

	res, err := s.txService.CreateDirectTransaction(ctx, req)
	if err != nil {
		slog.Error("Failed to process /catat via Telegram", slog.Any("error", err))
		return "❌ Gagal mencatat transaksi: saldo tidak cukup atau kesalahan sistem.", nil
	}

	return fmt.Sprintf("✅ Transaksi tersimpan! ID: `%s`, Nominal: Rp %.2f", res.ID, res.Amount), nil
}

func (s *BotService) handleSaldo(ctx context.Context, update *TelegramUpdate) (string, error) {
	chatID := update.Message.Chat.ID
	fam, err := s.familyService.FindByTelegramChatID(ctx, chatID)
	if err != nil || fam == nil {
		return "⚠️ Chat Telegram ini belum terhubung dengan keluarga. Gunakan `/link [invite_code]` terlebih dahulu.", nil
	}

	balances, err := s.familyService.GetWalletBalances(ctx, fam.ID)
	if err != nil {
		return "⚠️ Gagal mengambil saldo dompet.", nil
	}

	if len(balances) == 0 {
		return "Belum ada dompet terdaftar.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 *Ringkasan Saldo Dompet Keluarga (%s)*\n\n", fam.Name))
	for _, b := range balances {
		sb.WriteString(fmt.Sprintf("• *%s* (ID: `%s`): Rp %.2f (Min: Rp %.2f)\n", b.WalletName, b.WalletID, b.CurrentBalance, b.MinimumLimit))
	}

	return sb.String(), nil
}
