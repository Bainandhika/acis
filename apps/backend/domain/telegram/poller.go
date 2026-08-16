package telegram

import (
	"context"
	"log/slog"
	"time"
)

type BotPoller struct {
	client     *Client
	botService *BotService
	stopChan   chan struct{}
	offset     int
}

func NewBotPoller(client *Client, botService *BotService) *BotPoller {
	return &BotPoller{
		client:     client,
		botService: botService,
		stopChan:   make(chan struct{}),
		offset:     0,
	}
}

func (p *BotPoller) Start(ctx context.Context) {
	if p.client.botToken == "" {
		slog.Warn("TELEGRAM_BOT_TOKEN not configured, skipping bot poller")
		return
	}

	// Delete existing webhook so long-polling receives updates
	_ = p.client.DeleteWebhook(ctx)
	slog.Info("Telegram Bot Poller started (listening for incoming messages/commands)")

	go func() {
		for {
			select {
			case <-p.stopChan:
				return
			default:
				updates, err := p.client.GetUpdates(ctx, p.offset, 15)
				if err != nil {
					slog.Debug("Telegram getUpdates error", slog.Any("error", err))
					time.Sleep(2 * time.Second)
					continue
				}

				for _, u := range updates {
					if u.UpdateID >= p.offset {
						p.offset = u.UpdateID + 1
					}

					reply, err := p.botService.ProcessUpdate(ctx, &u)
					if err != nil {
						slog.Error("Failed to process telegram update", slog.Any("error", err), slog.Int("update_id", u.UpdateID))
						continue
					}

					if reply != "" && u.Message != nil {
						if err := p.client.SendMessage(ctx, u.Message.Chat.ID, reply); err != nil {
							slog.Error("Failed to send telegram reply", slog.Any("error", err), slog.Int64("chat_id", u.Message.Chat.ID))
						}
					}
				}
			}
		}
	}()
}

func (p *BotPoller) Stop() {
	close(p.stopChan)
	slog.Info("Telegram Bot Poller stopped")
}
