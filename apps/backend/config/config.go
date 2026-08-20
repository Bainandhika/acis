package config

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	AppDSN   string `yaml:"app_dsn"`
	AdminDSN string `yaml:"admin_dsn"`
}

type SupabaseConfig struct {
	JWKSURL string `yaml:"jwks_url"`
}

type LogConfig struct {
	Dir   string `yaml:"dir"`
	Level string `yaml:"level"`
}

type RedisConfig struct {
	URL      string `yaml:"url"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type TelegramConfig struct {
	BotToken      string `yaml:"bot_token"`
	WebhookSecret string `yaml:"webhook_secret"`
	BotUsername   string `yaml:"bot_username"`
}

type BotConfig struct {
	Secret string `yaml:"secret"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Supabase SupabaseConfig `yaml:"supabase"`
	Log      LogConfig      `yaml:"log"`
	Redis    RedisConfig    `yaml:"redis"`
	Telegram TelegramConfig `yaml:"telegram"`
	Bot      BotConfig      `yaml:"bot"`
	CORS     CORSConfig     `yaml:"cors"`
}

var envPattern = regexp.MustCompile(`\$\{([A-Za-z0-9_]+)(?::([^}]*))?\}`)

// expandEnv replaces ${VAR:default} or ${VAR} in string with environment variable or default value
func expandEnv(s string) string {
	return envPattern.ReplaceAllStringFunc(s, func(match string) string {
		submatches := envPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		varName := submatches[1]
		defaultVal := ""
		if len(submatches) >= 3 {
			defaultVal = submatches[2]
		}
		if val, ok := os.LookupEnv(varName); ok && val != "" {
			return val
		}
		return defaultVal
	})
}

// Load reads .env and acis-config.yaml, expanding environment variables
func Load(configPath string) (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Notice: .env file not found or skipped, using system environment variables")
	}

	if configPath == "" {
		configPath = "acis-config.yaml"
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Warning: Failed to read %s (%v)", configPath, err)
	}

	expandedContent := expandEnv(string(content))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expandedContent), &cfg); err != nil {
		return nil, fmt.Errorf("Warning: Failed to parse YAML config (%v)", err)
	}

	return &cfg, nil
}

// DSN is intentionally kept as an empty fallback to force explicit app_dsn/admin_dsn configuration.
func (c *Config) DSN() string {
	return ""
}

// AppDSN returns the DSN for user-scoped operations with RLS.
func (c *Config) AppDSN() string {
	return c.Database.AppDSN
}

// AdminDSN returns the DSN for admin / internal operations (bypassing RLS).
func (c *Config) AdminDSN() string {
	return c.Database.AdminDSN
}
