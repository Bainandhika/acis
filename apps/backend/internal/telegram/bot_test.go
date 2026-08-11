package telegram

import (
	"context"
	"strings"
	"testing"
)

func TestTelegramCommandHandler(t *testing.T) {
	handler := NewCommandHandler()
	ctx := context.Background()

	t.Run("Empty update or text", func(t *testing.T) {
		res, err := handler.ProcessMessage(ctx, &TelegramUpdate{})
		if err != nil || res != "" {
			t.Errorf("Expected empty response, got %s, err: %v", res, err)
		}
	})

	t.Run("/catat valid command", func(t *testing.T) {
		update := &TelegramUpdate{
			Message: &TelegramMessage{
				Text: "/catat Makan 50000 Nasi Goreng",
			},
		}
		res, err := handler.ProcessMessage(ctx, update)
		if err != nil || !strings.Contains(res, "Catatan tersimpan") {
			t.Errorf("Expected success response, got: %s", res)
		}
	})

	t.Run("/catat invalid format", func(t *testing.T) {
		update := &TelegramUpdate{
			Message: &TelegramMessage{
				Text: "/catat Makan",
			},
		}
		res, err := handler.ProcessMessage(ctx, update)
		if err != nil || !strings.Contains(res, "Format salah") {
			t.Errorf("Expected format error response, got: %s", res)
		}
	})

	t.Run("/saldo command", func(t *testing.T) {
		update := &TelegramUpdate{
			Message: &TelegramMessage{
				Text: "/saldo",
			},
		}
		res, err := handler.ProcessMessage(ctx, update)
		if err != nil || !strings.Contains(res, "Cek Saldo") {
			t.Errorf("Expected saldo response, got: %s", res)
		}
	})

	t.Run("Unknown command", func(t *testing.T) {
		update := &TelegramUpdate{
			Message: &TelegramMessage{
				Text: "/unknown",
			},
		}
		res, err := handler.ProcessMessage(ctx, update)
		if err != nil || !strings.Contains(res, "Perintah tidak dikenali") {
			t.Errorf("Expected unknown command response, got: %s", res)
		}
	})
}
