package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
)

type LowBalanceWorker struct {
	familyService FamilyService
	outboxRepo    notification.OutboxRepository
	db            *database.AppDB
	interval      time.Duration
	stopChan      chan struct{}
	ticker        *time.Ticker
}

func NewLowBalanceWorker(familyService FamilyService, outboxRepo notification.OutboxRepository, db *database.AppDB, interval time.Duration) *LowBalanceWorker {
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	return &LowBalanceWorker{
		familyService: familyService,
		outboxRepo:    outboxRepo,
		db:            db,
		interval:      interval,
		stopChan:      make(chan struct{}),
	}
}

func (w *LowBalanceWorker) Start(ctx context.Context) {
	w.ticker = time.NewTicker(w.interval)
	slog.Info("Telegram low-balance reminder worker started", slog.Duration("interval", w.interval))

	go func() {
		for {
			select {
			case <-w.stopChan:
				return
			case <-w.ticker.C:
				w.checkAndEnqueueReminders(ctx)
			}
		}
	}()
}

func (w *LowBalanceWorker) checkAndEnqueueReminders(ctx context.Context) {
	wallets, err := w.familyService.GetLowBalanceWallets(ctx)
	if err != nil {
		slog.Error("Failed to check low-balance wallets", slog.Any("error", err))
		return
	}

	if len(wallets) == 0 {
		return
	}

	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("Failed to begin transaction for low balance alerts", slog.Any("error", err))
		return
	}
	defer tx.Rollback()

	for _, wallet := range wallets {
		recipient := fmt.Sprintf("family:%s", wallet.FamilyID)
		if wallet.TelegramChatID != nil {
			recipient = fmt.Sprintf("%d", *wallet.TelegramChatID)
		}

		payload := map[string]interface{}{
			"wallet_id":       wallet.WalletID,
			"wallet_name":     wallet.WalletName,
			"current_balance": wallet.CurrentBalance,
			"minimum_limit":   wallet.MinimumLimit,
			"alert_type":      "low_balance",
		}

		if err := w.outboxRepo.EnqueueTx(ctx, tx, "telegram_alert", recipient, payload); err != nil {
			slog.Error("Failed to enqueue low-balance alert to outbox", slog.Any("error", err), slog.String("wallet_id", wallet.WalletID))
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit low-balance alert outbox transaction", slog.Any("error", err))
	} else {
		_ = w.outboxRepo.PublishSignal(ctx)
		slog.Info("Low-balance outbox alerts enqueued successfully", slog.Int("alerts_enqueued", len(wallets)))
	}
}

func (w *LowBalanceWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	select {
	case <-w.stopChan:
	default:
		close(w.stopChan)
	}
	slog.Info("Low-balance reminder worker stopped")
}
