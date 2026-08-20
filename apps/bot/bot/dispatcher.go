package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Bainandhika/acis/apps/bot/backendclient"
	"github.com/Bainandhika/acis/apps/bot/telegram"
)

const startPrompt = "Ketik /start untuk memulai bot ACIS."

const onboardingMessage = `👋 *Selamat datang di ACIS Bot!*

Saya membantu Anda mencatat transaksi keuangan, menerima notifikasi, dan melihat saldo dompet keluarga langsung dari Telegram.

📋 *Daftar Perintah:*
• /link {kode} | /hubungkan {kode}
  _Menghubungkan akun Telegram Anda dengan akun ACIS atau grup ke keluarga._
• /balance | /saldo
  _Melihat saldo seluruh dompet keluarga beserta ID Dompet._
• /transaction {id dompet} {pemasukan/pengeluaran} {nominal} {keterangan} | /transaksi {id dompet} {pemasukan/pengeluaran} {nominal} {keterangan}
  _Mencatat transaksi baru ke dompet keluarga._

💡 *Contoh:*
• /link ABC123
• /saldo
• /transaksi SMTH01-1 pengeluaran 50000 Belanja sayur di pasar`

type Dispatcher struct {
	tg *telegram.Client
	bc *backendclient.Client
}

func NewDispatcher(tg *telegram.Client, bc *backendclient.Client) *Dispatcher {
	return &Dispatcher{
		tg: tg,
		bc: bc,
	}
}

func (d *Dispatcher) ProcessUpdate(ctx context.Context, u *telegram.Update) error {
	// Handle user leaving or kicking the bot from chat / Telegram
	if u.MyChatMember != nil {
		status := u.MyChatMember.NewMember.Status
		if status == "left" || status == "kicked" {
			// Bot session terminated (stateless client, no persistent session to drop)
			return nil
		}
	}

	if u.Message == nil || u.Message.Text == "" {
		return nil
	}

	text := strings.TrimSpace(u.Message.Text)
	chatID := u.Message.Chat.ID

	if !strings.HasPrefix(text, "/") {
		return d.tg.SendMessage(ctx, chatID, startPrompt)
	}

	parts := strings.Fields(text)
	cmd := strings.ToLower(strings.Split(parts[0], "@")[0])

	switch cmd {
	case "/start":
		return d.tg.SendMessage(ctx, chatID, onboardingMessage)
	case "/link", "/hubungkan":
		return d.handleLink(ctx, chatID, parts)
	case "/balance", "/saldo":
		return d.handleBalance(ctx, chatID)
	case "/transaction", "/transaksi":
		return d.handleTransaction(ctx, chatID, parts)
	default:
		return d.tg.SendMessage(ctx, chatID, startPrompt)
	}
}

func (d *Dispatcher) handleLink(ctx context.Context, chatID int64, parts []string) error {
	if len(parts) < 2 {
		return d.tg.SendMessage(ctx, chatID, "❌ Format salah.\nGunakan: `/link [kode]` atau `/hubungkan [kode]`\nContoh: `/link ABC123`")
	}

	code := strings.ToUpper(strings.TrimSpace(parts[1]))

	// First try linking user account (Supabase profile linking via 6-character code)
	errUser := d.bc.LinkUserAccount(ctx, code, chatID)
	if errUser == nil {
		return d.tg.SendMessage(ctx, chatID, fmt.Sprintf("✅ Berhasil menghubungkan akun Telegram Anda ke profil ACIS!\n\nKode verifikasi: `%s`", code))
	}

	// Fallback: try linking family invite code (if connecting group or family chat)
	if errFam := d.bc.LinkFamily(ctx, code, chatID); errFam == nil {
		return d.tg.SendMessage(ctx, chatID, fmt.Sprintf("✅ Berhasil terhubung ke keluarga dengan kode undangan `%s`!\n\nGunakan /saldo atau /balance untuk melihat daftar dompet.", code))
	}

	return d.tg.SendMessage(ctx, chatID, fmt.Sprintf("❌ Gagal menghubungkan: %s", errUser.Error()))
}

func (d *Dispatcher) handleBalance(ctx context.Context, chatID int64) error {
	fam, err := d.bc.GetFamily(ctx, chatID)
	if err != nil || fam == nil {
		return d.tg.SendMessage(ctx, chatID, "⚠️ Chat ini belum terhubung ke keluarga manapun.\nGunakan `/link [kode_undangan]` terlebih dahulu.")
	}

	balances, err := d.bc.GetBalance(ctx, chatID)
	if err != nil {
		return d.tg.SendMessage(ctx, chatID, fmt.Sprintf("⚠️ Gagal mengambil saldo: %s", err.Error()))
	}

	if len(balances) == 0 {
		return d.tg.SendMessage(ctx, chatID, fmt.Sprintf("Belum ada dompet terdaftar di keluarga *%s*.", fam.Name))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 *Saldo Dompet Keluarga — %s*\n\n", fam.Name))
	for _, b := range balances {
		sb.WriteString(fmt.Sprintf("• *%s* (ID: `%s`)\n  Saldo: Rp %.2f (Batas Min: Rp %.2f)\n\n",
			b.WalletName, b.ShortID, b.CurrentBalance, b.MinimumLimit))
	}
	sb.WriteString("💡 _Gunakan ID Dompet di atas untuk mencatat transaksi dengan /transaksi._")

	return d.tg.SendMessage(ctx, chatID, sb.String())
}

func (d *Dispatcher) handleTransaction(ctx context.Context, chatID int64, parts []string) error {
	if len(parts) < 4 {
		return d.tg.SendMessage(ctx, chatID,
			"❌ Format salah.\n"+
				"Gunakan: `/transaksi [wallet_id] [pemasukan/pengeluaran] [nominal] [keterangan]`\n"+
				"Atau: `/transaction [wallet_id] [income/expense] [amount] [description]`\n\n"+
				"Contoh:\n`/transaksi SMTH01-1 pengeluaran 50000 Beli beras`\n`/transaction SMTH01-1 income 1000000 Bonus bulanan`")
	}

	walletShortID := parts[1]

	rawType := strings.ToLower(parts[2])
	var txType string
	switch rawType {
	case "income", "pemasukan", "masuk":
		txType = "income"
	case "expense", "pengeluaran", "keluar":
		txType = "expense"
	default:
		return d.tg.SendMessage(ctx, chatID, "❌ Jenis transaksi tidak valid. Gunakan `pemasukan`/`income` atau `pengeluaran`/`expense`.")
	}

	cleanAmountStr := strings.ReplaceAll(parts[3], ".", "")
	cleanAmountStr = strings.ReplaceAll(cleanAmountStr, ",", ".")
	cleanAmountStr = strings.TrimPrefix(cleanAmountStr, "rp")
	cleanAmountStr = strings.TrimPrefix(cleanAmountStr, "Rp")

	amount, err := strconv.ParseFloat(cleanAmountStr, 64)
	if err != nil || amount <= 0 {
		return d.tg.SendMessage(ctx, chatID, "❌ Nominal tidak valid. Masukkan angka nominal yang benar.")
	}

	description := ""
	if len(parts) >= 5 {
		description = strings.Join(parts[4:], " ")
	}

	res, err := d.bc.RecordTransaction(ctx, backendclient.RecordTxReq{
		WalletShortID: walletShortID,
		Type:          txType,
		Amount:        amount,
		Description:   description,
		ChatID:        chatID,
	})
	if err != nil {
		return d.tg.SendMessage(ctx, chatID, fmt.Sprintf("❌ Gagal mencatat transaksi: %s", err.Error()))
	}

	typeLabel := "Pengeluaran (-)"
	if txType == "income" {
		typeLabel = "Pemasukan (+)"
	}

	descText := "-"
	if description != "" {
		descText = description
	}

	return d.tg.SendMessage(ctx, chatID,
		fmt.Sprintf("✅ *Transaksi Berhasil Dicatat!*\n\n"+
			"• Dompet: `%s`\n"+
			"• Jenis: *%s*\n"+
			"• Nominal: *Rp %.2f*\n"+
			"• Keterangan: %s\n"+
			"• ID Transaksi: `%s`",
			walletShortID, typeLabel, amount, descText, res.ID))
}
