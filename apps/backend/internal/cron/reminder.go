package cron

import (
	"context"
	"time"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/rs/zerolog/log"
)

type BalanceReminderWorker struct {
	db       *database.AppDB
	interval time.Duration
}

func NewBalanceReminderWorker(db *database.AppDB, interval time.Duration) *BalanceReminderWorker {
	return &BalanceReminderWorker{
		db:       db,
		interval: interval,
	}
}

func (w *BalanceReminderWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Balance reminder worker stopping...")
			return
		case <-ticker.C:
			w.checkLowBalanceWallets(ctx)
		}
	}
}

func (w *BalanceReminderWorker) checkLowBalanceWallets(ctx context.Context) {
	query := `SELECT id, family_id, name, current_balance, minimum_limit 
	          FROM wallets 
	          WHERE current_balance <= minimum_limit AND minimum_limit > 0`

	var lowBalanceWallets []domain.Wallet
	err := w.db.SelectContext(ctx, &lowBalanceWallets, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query low balance wallets")
		return
	}

	for _, wallet := range lowBalanceWallets {
		log.Warn().
			Str("wallet_id", wallet.ID).
			Str("wallet_name", wallet.Name).
			Float64("current_balance", wallet.CurrentBalance).
			Float64("minimum_limit", wallet.MinimumLimit).
			Msg("⚠️ WARNING: Wallet balance reached or fell below minimum limit!")
	}
}
