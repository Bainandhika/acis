package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandEnv(t *testing.T) {
	os.Setenv("TEST_VAR_1", "custom_value")
	defer os.Unsetenv("TEST_VAR_1")

	input := `host: "${TEST_VAR_1}"
port: "${TEST_VAR_UNSET:5432}"`

	result := expandEnv(input)
	expected := `host: "custom_value"
port: "5432"`

	if result != expected {
		t.Fatalf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "acis-config.yaml")

	content := `server:
  port: "${TEST_PORT:9090}"
database:
  app_dsn: "postgresql://app_user:app_pass@127.0.0.1:6543/postgres?sslmode=require"
  admin_dsn: "postgresql://admin_user:admin_pass@127.0.0.1:5432/postgres?sslmode=require"
supabase:
  jwks_url: "https://example.supabase.co/auth/v1/.well-known/jwks.json"
telegram:
  bot_token: "test_bot_token"
  bot_username: "acis_bot"
bot:
  secret: "test_internal_secret"`

	if err := os.WriteFile(yamlFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg, err := Load(yamlFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("expected Server.Port 9090, got %s", cfg.Server.Port)
	}
	if cfg.Database.AppDSN != "postgresql://app_user:app_pass@127.0.0.1:6543/postgres?sslmode=require" {
		t.Errorf("expected Database.AppDSN matching config, got %s", cfg.Database.AppDSN)
	}
	if cfg.Database.AdminDSN != "postgresql://admin_user:admin_pass@127.0.0.1:5432/postgres?sslmode=require" {
		t.Errorf("expected Database.AdminDSN matching config, got %s", cfg.Database.AdminDSN)
	}
	if cfg.Supabase.JWKSURL != "https://example.supabase.co/auth/v1/.well-known/jwks.json" {
		t.Errorf("expected Supabase.JWKSURL matching config, got %s", cfg.Supabase.JWKSURL)
	}
	if cfg.Telegram.BotToken != "test_bot_token" {
		t.Errorf("expected Telegram.BotToken 'test_bot_token', got %s", cfg.Telegram.BotToken)
	}
	if cfg.Telegram.BotUsername != "acis_bot" {
		t.Errorf("expected Telegram.BotUsername 'acis_bot', got %s", cfg.Telegram.BotUsername)
	}
	if cfg.Bot.Secret != "test_internal_secret" {
		t.Errorf("expected Bot.Secret 'test_internal_secret', got %s", cfg.Bot.Secret)
	}
}

func TestLoadConfig_TelegramEnv(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "env_token_123")
	os.Setenv("TELEGRAM_WEBHOOK_SECRET", "env_secret_456")
	os.Setenv("TELEGRAM_BOT_USERNAME", "env_bot")
	defer func() {
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("TELEGRAM_WEBHOOK_SECRET")
		os.Unsetenv("TELEGRAM_BOT_USERNAME")
	}()

	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "acis-config.yaml")

	content := `server:
  port: "8080"
telegram:
  bot_token: "${TELEGRAM_BOT_TOKEN:}"
  webhook_secret: "${TELEGRAM_WEBHOOK_SECRET:}"
  bot_username: "${TELEGRAM_BOT_USERNAME:}"`

	if err := os.WriteFile(yamlFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg, err := Load(yamlFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Telegram.BotToken != "env_token_123" {
		t.Errorf("expected Telegram.BotToken 'env_token_123', got %s", cfg.Telegram.BotToken)
	}
	if cfg.Telegram.WebhookSecret != "env_secret_456" {
		t.Errorf("expected Telegram.WebhookSecret 'env_secret_456', got %s", cfg.Telegram.WebhookSecret)
	}
	if cfg.Telegram.BotUsername != "env_bot" {
		t.Errorf("expected Telegram.BotUsername 'env_bot', got %s", cfg.Telegram.BotUsername)
	}
}

func TestDatabaseDSNUsesOnlyAppAndAdminDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{},
	}

	if got := cfg.AppDSN(); got != "" {
		t.Fatalf("expected empty app DSN when app_dsn is not configured, got %q", got)
	}
	if got := cfg.AdminDSN(); got != "" {
		t.Fatalf("expected empty admin DSN when admin_dsn is not configured, got %q", got)
	}
	if got := cfg.DSN(); got != "" {
		t.Fatalf("expected empty legacy DSN fallback, got %q", got)
	}

	cfg.Database.AppDSN = "postgresql://app_user:app_pass@host:5432/app_db?sslmode=require"
	cfg.Database.AdminDSN = "postgresql://admin_user:admin_pass@host:5432/postgres?sslmode=require"
	if got := cfg.AppDSN(); got != cfg.Database.AppDSN {
		t.Fatalf("expected app DSN to match configured app_dsn, got %q", got)
	}
	if got := cfg.AdminDSN(); got != cfg.Database.AdminDSN {
		t.Fatalf("expected admin DSN to match configured admin_dsn, got %q", got)
	}
}
