package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/resend/resend-go/v2"
)

func TestNewResendSender_DefaultSender(t *testing.T) {
	// Test without 'from' argument
	sender1 := NewResendSender("re_test_123")
	if sender1.From() != DefaultSenderAddress {
		t.Fatalf("expected default sender %q, got %q", DefaultSenderAddress, sender1.From())
	}

	// Test with empty string 'from'
	sender2 := NewResendSender("re_test_123", "")
	if sender2.From() != DefaultSenderAddress {
		t.Fatalf("expected default sender %q, got %q", DefaultSenderAddress, sender2.From())
	}
}

func TestNewResendSender_CustomSender(t *testing.T) {
	customSender := "ACIS Support <support@acis.id>"
	sender := NewResendSender("re_test_123", customSender)
	if sender.From() != customSender {
		t.Fatalf("expected custom sender %q, got %q", customSender, sender.From())
	}
}

func TestResendSender_SendOTP_NoApiKey(t *testing.T) {
	sender := NewResendSender("")
	err := sender.SendOTP(context.Background(), "user@example.com", "123456")
	if err != nil {
		t.Fatalf("expected no error when api key is not configured, got: %v", err)
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestResendSender_SendOTP_Success(t *testing.T) {
	var capturedReq struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Html    string   `json:"html"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer re_test_key" {
			http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "msg_12345"}`))
	}))
	defer server.Close()

	// Redirect HTTP requests to mock test server
	customHttpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			targetUrl := server.URL + req.URL.Path
			newReq, err := http.NewRequestWithContext(req.Context(), req.Method, targetUrl, req.Body)
			if err != nil {
				return nil, err
			}
			newReq.Header = req.Header
			return http.DefaultTransport.RoundTrip(newReq)
		}),
	}

	client := resend.NewCustomClient(customHttpClient, "re_test_key")
	sender := NewResendSenderWithClient(client, "ACIS <noreply@acis.id>")

	err := sender.SendOTP(context.Background(), "testuser@domain.com", "789012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.From != "ACIS <noreply@acis.id>" {
		t.Errorf("expected from 'ACIS <noreply@acis.id>', got %q", capturedReq.From)
	}
	if len(capturedReq.To) != 1 || capturedReq.To[0] != "testuser@domain.com" {
		t.Errorf("expected to 'testuser@domain.com', got %v", capturedReq.To)
	}
	if capturedReq.Subject != "Kode OTP ACIS Anda" {
		t.Errorf("expected subject 'Kode OTP ACIS Anda', got %q", capturedReq.Subject)
	}
	if !strings.Contains(capturedReq.Html, "789012") {
		t.Errorf("expected html to contain OTP code '789012', got %q", capturedReq.Html)
	}
}

func TestResendSender_SendOTP_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"statusCode": 400, "message": "Invalid email address", "name": "validation_error"}`))
	}))
	defer server.Close()

	customHttpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			targetUrl := server.URL + req.URL.Path
			newReq, err := http.NewRequestWithContext(req.Context(), req.Method, targetUrl, req.Body)
			if err != nil {
				return nil, err
			}
			newReq.Header = req.Header
			return http.DefaultTransport.RoundTrip(newReq)
		}),
	}

	client := resend.NewCustomClient(customHttpClient, "re_test_key")
	sender := NewResendSenderWithClient(client, "")

	if sender.From() != DefaultSenderAddress {
		t.Errorf("expected default sender, got %q", sender.From())
	}

	err := sender.SendOTP(context.Background(), "invalid-email", "123456")
	if err == nil {
		t.Fatal("expected error from API response, got nil")
	}

	if !strings.Contains(err.Error(), "failed to send OTP email via Resend") {
		t.Errorf("expected error message to contain 'failed to send OTP email via Resend', got: %v", err)
	}
}
