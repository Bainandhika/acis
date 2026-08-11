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
  host: "127.0.0.1"
  password: "secret_db_pass"`

	if err := os.WriteFile(yamlFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg := Load(yamlFile)
	if cfg.Server.Port != "9090" {
		t.Errorf("expected Server.Port 9090, got %s", cfg.Server.Port)
	}
	if cfg.Database.Host != "127.0.0.1" {
		t.Errorf("expected Database.Host 127.0.0.1, got %s", cfg.Database.Host)
	}
	if cfg.Database.Password != "secret_db_pass" {
		t.Errorf("expected Database.Password secret_db_pass, got %s", cfg.Database.Password)
	}
}
