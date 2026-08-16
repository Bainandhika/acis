package notification

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v2"
)

const DefaultSenderAddress = "onboarding@resend.dev"

// EmailSender defines the interface for sending notification emails.
type EmailSender interface {
	SendOTP(ctx context.Context, email, code string) error
}

// ResendSender sends emails using the official Resend Go SDK.
type ResendSender struct {
	client *resend.Client
	apiKey string
	from   string
}

// NewResendSender creates a new ResendSender instance.
// If from is empty, it defaults to DefaultSenderAddress ("onboarding@resend.dev").
func NewResendSender(apiKey string, from ...string) *ResendSender {
	senderFrom := DefaultSenderAddress
	if len(from) > 0 && from[0] != "" {
		senderFrom = from[0]
	}

	var client *resend.Client
	if apiKey != "" {
		client = resend.NewClient(apiKey)
	}

	return &ResendSender{
		client: client,
		apiKey: apiKey,
		from:   senderFrom,
	}
}

// NewResendSenderWithClient creates a ResendSender with a pre-configured Resend client (e.g. for testing).
func NewResendSenderWithClient(client *resend.Client, from string) *ResendSender {
	if from == "" {
		from = DefaultSenderAddress
	}
	return &ResendSender{
		client: client,
		from:   from,
	}
}

// From returns the configured sender email address.
func (s *ResendSender) From() string {
	return s.from
}

// SendOTP dispatches an OTP verification email using the Resend Go SDK.
func (s *ResendSender) SendOTP(ctx context.Context, email, code string) error {
	if s.client == nil {
		slog.Warn("Resend API key not configured, skipping real email dispatch", slog.String("recipient", email))
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{email},
		Subject: "Kode OTP ACIS Anda",
		Html:    fmt.Sprintf("<p>Kode OTP ACIS Anda adalah: <b>%s</b></p><p>Kode berlaku selama 5 menit.</p>", code),
	}

	resp, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send OTP email via Resend: %w", err)
	}

	slog.Info("OTP email sent successfully via Resend",
		slog.String("recipient", email),
		slog.String("email_id", resp.Id),
		slog.String("from", s.from),
	)
	return nil
}
