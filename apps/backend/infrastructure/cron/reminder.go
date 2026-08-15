package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
)

type WalletMinBalance struct {
	ID             string  `db:"id"`
	FamilyID       string  `db:"family_id"`
	Name           string  `db:"name"`
	CurrentBalance float64 `db:"current_balance"`
	MinimumLimit   float64 `db:"minimum_limit"`
}

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
			slog.Info("Balance reminder worker stopping...")
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

	var lowBalanceWallets []WalletMinBalance
	err := w.db.SelectContext(ctx, &lowBalanceWallets, query)
	if err != nil {
		slog.Error("Failed to query low balance wallets", slog.Any("error", err))
		return
	}

	for _, wallet := range lowBalanceWallets {
		slog.Warn("WARNING: Wallet balance reached or fell below minimum limit!",
			slog.String("wallet_id", wallet.ID),
			slog.String("wallet_name", wallet.Name),
			slog.Float64("current_balance", wallet.CurrentBalance),
			slog.Float64("minimum_limit", wallet.MinimumLimit),
		)
	}
}
