package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	botToken   string
	httpClient *http.Client
}

type GetUpdatesResponse struct {
	OK          bool             `json:"ok"`
	Result      []TelegramUpdate `json:"result"`
	Description string           `json:"description,omitempty"`
}

func NewClient(botToken string) *Client {
	return &Client{
		botToken: botToken,
		httpClient: &http.Client{
			Timeout: 35 * time.Second, // Allow enough timeout for long polling (e.g. 20s timeout)
		},
	}
}

func (c *Client) DeleteWebhook(ctx context.Context) error {
	if c.botToken == "" {
		return nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook?drop_pending_updates=false", c.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

func (c *Client) GetUpdates(ctx context.Context, offset int, timeout int) ([]TelegramUpdate, error) {
	if c.botToken == "" {
		return nil, nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=%d", c.botToken, offset, timeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create getUpdates request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram API returned status: %d", resp.StatusCode)
	}

	var updateResp GetUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return nil, fmt.Errorf("failed to decode getUpdates response: %w", err)
	}

	if !updateResp.OK {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", updateResp.Description)
	}

	return updateResp.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	if c.botToken == "" {
		slog.Warn("Telegram bot token not configured, skipping SendMessage")
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	// Retry up to 3 times for transient failures
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create telegram request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt*200) * time.Millisecond)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			return nil
		}

		_ = resp.Body.Close()
		lastErr = fmt.Errorf("telegram API returned status: %d", resp.StatusCode)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Client-side errors (invalid chat_id or token) should not be retried
			break
		}
		time.Sleep(time.Duration(attempt*200) * time.Millisecond)
	}

	return lastErr
}
