package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bainandhika/acis/apps/bot/backendclient"
	botpkg "github.com/Bainandhika/acis/apps/bot/bot"
	"github.com/Bainandhika/acis/apps/bot/telegram"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	backendURL := os.Getenv("BACKEND_BASE_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	botSecret := os.Getenv("BOT_INTERNAL_SECRET")

	tg := telegram.NewClient(botToken)
	bc := backendclient.New(backendURL, botSecret)
	d := botpkg.NewDispatcher(tg, bc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = tg.DeleteWebhook(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	offset := 0
	slog.Info("ACIS Bot service started. Polling Telegram for updates...", slog.String("backend_url", backendURL))

	for {
		select {
		case <-quit:
			slog.Info("ACIS Bot service shutting down...")
			return
		default:
			updates, err := tg.GetUpdates(ctx, offset, 20)
			if err != nil {
				slog.Warn("Telegram getUpdates error", slog.Any("error", err))
				time.Sleep(2 * time.Second)
				continue
			}

			for i := range updates {
				u := &updates[i]
				if u.UpdateID >= offset {
					offset = u.UpdateID + 1
				}
				if err := d.ProcessUpdate(ctx, u); err != nil {
					slog.Error("ProcessUpdate error", slog.Any("error", err), slog.Int("update_id", u.UpdateID))
				}
			}
		}
	}
}
