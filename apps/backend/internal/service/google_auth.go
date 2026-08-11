package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type GoogleTokenPayload struct {
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"` // Google ID
}

// VerifyGoogleIDToken verifies a Google OAuth 2.0 ID Token using Google's TokenInfo API endpoint
func VerifyGoogleIDToken(ctx context.Context, idToken string) (*GoogleTokenPayload, error) {
	if idToken == "" {
		return nil, errors.New("google id_token cannot be empty")
	}

	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify google token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid google id_token")
	}

	var payload GoogleTokenPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode google token response: %w", err)
	}

	return &payload, nil
}
