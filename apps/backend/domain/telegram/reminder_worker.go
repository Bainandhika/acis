package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	zerolog "github.com/rs/zerolog/log"
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
	zerolog.Info().Dur("interval", w.interval).Msg("Telegram low-balance reminder worker started")

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
		zerolog.Error().Err(err).Msg("Failed to check low-balance wallets")
		return
	}

	if len(wallets) == 0 {
		return
	}

	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		zerolog.Error().Err(err).Msg("Failed to begin transaction for low balance alerts")
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
			zerolog.Error().Err(err).Str("wallet_id", wallet.WalletID).Msg("Failed to enqueue low-balance alert to outbox")
		}
	}

	if err := tx.Commit(); err != nil {
		zerolog.Error().Err(err).Msg("Failed to commit low-balance alert outbox transaction")
	} else {
		zerolog.Info().Int("alerts_enqueued", len(wallets)).Msg("Low-balance outbox alerts enqueued successfully")
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
	zerolog.Info().Msg("Low-balance reminder worker stopped")
}
