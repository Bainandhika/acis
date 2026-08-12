package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	zerolog "github.com/rs/zerolog/log"
)

type ResendSender struct {
	apiKey     string
	httpClient *http.Client
}

func NewResendSender(apiKey string) *ResendSender {
	return &ResendSender{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *ResendSender) SendOTP(ctx context.Context, email, code string) error {
	if s.apiKey == "" {
		zerolog.Warn().Str("recipient", email).Msg("RESEND_API_KEY not configured, skipping real email dispatch")
		return nil
	}

	url := "https://api.resend.com/emails"
	payload := map[string]interface{}{
		"from":    "ACIS App <noreply@acis.app>",
		"to":      []string{email},
		"subject": "Kode OTP ACIS Anda",
		"html":    fmt.Sprintf("<p>Kode OTP ACIS Anda adalah: <b>%s</b></p><p>Kode berlaku selama 5 menit.</p>", code),
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API returned error status: %d", resp.StatusCode)
	}

	return nil
}
