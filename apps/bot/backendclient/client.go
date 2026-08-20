package backendclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	botSecret  string
	httpClient *http.Client
}

type Family struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	InviteCode     string `json:"invite_code"`
	TelegramChatID *int64 `json:"telegram_chat_id,omitempty"`
	WalletCounter  int    `json:"wallet_counter"`
}

type WalletBalance struct {
	WalletID       string  `json:"wallet_id"`
	ShortID        string  `json:"short_id"`
	WalletName     string  `json:"wallet_name"`
	CurrentBalance float64 `json:"current_balance"`
	MinimumLimit   float64 `json:"minimum_limit"`
}

type RecordTxReq struct {
	WalletShortID string  `json:"wallet_short_id"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	ChatID        int64   `json:"chat_id"`
}

type TxResult struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	UserID      *string   `json:"user_id,omitempty"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func New(baseURL, botSecret string) *Client {
	return &Client{
		baseURL:   baseURL,
		botSecret: botSecret,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type apiResponse[T any] struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    T      `json:"data,omitempty"`
}

func (c *Client) LinkFamily(ctx context.Context, inviteCode string, chatID int64) error {
	payload := map[string]interface{}{
		"invite_code": inviteCode,
		"chat_id":     chatID,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/bot/link", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.botSecret != "" {
		req.Header.Set("X-Bot-Secret", c.botSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp apiResponse[any]
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetFamily(ctx context.Context, chatID int64) (*Family, error) {
	url := fmt.Sprintf("%s/api/v1/bot/family?chat_id=%d", c.baseURL, chatID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.botSecret != "" {
		req.Header.Set("X-Bot-Secret", c.botSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp apiResponse[any]
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	var res apiResponse[Family]
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res.Data, nil
}

func (c *Client) GetBalance(ctx context.Context, chatID int64) ([]WalletBalance, error) {
	url := fmt.Sprintf("%s/api/v1/bot/balance?chat_id=%d", c.baseURL, chatID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.botSecret != "" {
		req.Header.Set("X-Bot-Secret", c.botSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp apiResponse[any]
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	var res apiResponse[[]WalletBalance]
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (c *Client) RecordTransaction(ctx context.Context, txReq RecordTxReq) (*TxResult, error) {
	bodyBytes, err := json.Marshal(txReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/bot/transaction", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.botSecret != "" {
		req.Header.Set("X-Bot-Secret", c.botSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp apiResponse[any]
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	var res apiResponse[TxResult]
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res.Data, nil
}
