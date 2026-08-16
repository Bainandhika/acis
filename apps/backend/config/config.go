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
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
	TimeZone string `yaml:"time_zone"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expiry string `yaml:"expiry"`
}

type LogConfig struct {
	Dir   string `yaml:"dir"`
	Level string `yaml:"level"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type OTPConfig struct {
	TTL           string `yaml:"ttl"`
	EncryptionKey string `yaml:"encryption_key"`
}

type EmailConfig struct {
	APIKey string `yaml:"api_key"`
	From   string `yaml:"from"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
	Redis    RedisConfig    `yaml:"redis"`
	OTP      OTPConfig      `yaml:"otp"`
	Email    EmailConfig    `yaml:"email"`
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
func Load(configPath string) *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Notice: .env file not found or skipped, using system environment variables")
	}

	if configPath == "" {
		configPath = "acis-config.yaml"
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: Failed to read %s (%v), using default settings", configPath, err)
		return defaultFallback()
	}

	expandedContent := expandEnv(string(content))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expandedContent), &cfg); err != nil {
		log.Printf("Warning: Failed to parse YAML config (%v), using fallback", err)
		return defaultFallback()
	}

	if cfg.Email.From == "" {
		cfg.Email.From = "onboarding@resend.dev"
	}
	if cfg.Email.APIKey == "" {
		cfg.Email.APIKey = os.Getenv("RESEND_API_KEY")
	}

	return &cfg
}

func defaultFallback() *Config {
	return &Config{
		Server: ServerConfig{
			Port: os.Getenv("PORT"),
			Mode: os.Getenv("GIN_MODE"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  "disable",
			TimeZone: "Asia/Jakarta",
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
			Expiry: "24h",
		},
		Log: LogConfig{
			Dir:   "./logs",
			Level: "info",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
		OTP: OTPConfig{
			TTL:           "5m",
			EncryptionKey: os.Getenv("OTP_ENCRYPTION_KEY"),
		},
		Email: EmailConfig{
			APIKey: os.Getenv("RESEND_API_KEY"),
			From:   "onboarding@resend.dev",
		},
	}
}


// DSN returns Data Source Name string for database connection
func (c *Config) DSN() string {
	host := c.Database.Host
	if host == "" {
		host = "localhost"
	}
	port := c.Database.Port
	if port == "" {
		port = "5432"
	}
	user := c.Database.User
	if user == "" {
		user = "acis_user"
	}
	name := c.Database.Name
	if name == "" {
		name = "acis_db"
	}
	sslMode := c.Database.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	tz := c.Database.TimeZone
	if tz == "" {
		tz = "Asia/Jakarta"
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, port, user, c.Database.Password, name, sslMode, tz,
	)
}
