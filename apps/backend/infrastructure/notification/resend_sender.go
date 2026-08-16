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

// SendOTP dispatches an OTP verification email using the Resend Go SDK with branded HTML template.
func (s *ResendSender) SendOTP(ctx context.Context, email, code string) error {
	if s.client == nil {
		slog.Warn("Resend API key not configured, skipping real email dispatch", slog.String("recipient", email))
		return nil
	}

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; margin: 0; padding: 24px; color: #1e293b; }
    .container { max-width: 520px; margin: 0 auto; background: #ffffff; border-radius: 16px; border: 1px solid #e2e8f0; padding: 32px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
    .badge { display: inline-block; background: #0f172a; color: #ffffff; font-weight: 800; font-size: 14px; padding: 4px 12px; border-radius: 8px; margin-bottom: 16px; }
    h1 { font-size: 20px; font-weight: 800; margin: 0 0 12px; color: #0f172a; }
    p { font-size: 14px; line-height: 1.5; color: #475569; margin: 0 0 20px; }
    .otp-box { background: #f1f5f9; border: 2px dashed #cbd5e1; border-radius: 12px; text-align: center; padding: 18px; margin: 24px 0; }
    .otp-code { font-family: monospace; font-size: 32px; font-weight: 900; letter-spacing: 6px; color: #0f172a; }
    .footer { font-size: 12px; color: #94a3b8; text-align: center; margin-top: 28px; border-top: 1px solid #f1f5f9; padding-top: 16px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="badge">ACIS Security</div>
    <h1>Kode Verifikasi Masuk</h1>
    <p>Gunakan kode OTP 6-digit berikut untuk masuk ke akun ACIS Anda. Jangan bagikan kode ini kepada siapapun.</p>
    <div class="otp-box">
      <div class="otp-code">%s</div>
    </div>
    <p style="font-size: 12px; color: #64748b;">⏳ Kode verifikasi ini berlaku selama <b>5 menit</b>.</p>
    <div class="footer">
      ACIS (Aplikasi Catatan Keuangan Istri/Suami) &bull; Dilindungi enkripsi end-to-end
    </div>
  </div>
</body>
</html>`, code)

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{email},
		Subject: fmt.Sprintf("[%s] Kode OTP Masuk ACIS", code),
		Html:    htmlContent,
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
