package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Global rotator reference to allow clean closing if needed
var globalRotator *MonthlyRotator

// ParseLevel converts a string log level to slog.Level
func ParseLevel(lvl string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(lvl)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Init initializes the global slog logger with monthly file rotation
func Init(logDir string, levelStr ...string) {
	lvl := "info"
	if len(levelStr) > 0 && levelStr[0] != "" {
		lvl = levelStr[0]
	}
	logLevel := ParseLevel(lvl)

	if logDir == "" {
		logDir = "./logs"
	}
	_ = os.MkdirAll(logDir, 0755)

	rotator, err := NewMonthlyRotator(logDir, "acis.log")
	if err != nil {
		// Fallback to stderr if rotator fails
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true,
		})
		slog.SetDefault(slog.New(handler))
		slog.Error("Failed to initialize monthly log rotator, logging to stdout only", slog.Any("error", err))
		return
	}
	globalRotator = rotator

	handler := slog.NewJSONHandler(rotator, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(slog.TimeKey, t.Format(time.RFC3339))
				}
			}
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
					src.File = filepath.Base(src.File)
				}
			}
			return a
		},
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Info("Logger initialized successfully with monthly rotation", slog.String("configured_level", lvl))
}

// Close flushes and closes active log file
func Close() error {
	if globalRotator != nil {
		return globalRotator.Close()
	}
	return nil
}

// Helper wrapper functions
func Debug(msg string, args ...any) {
	slog.Default().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Default().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Default().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Default().Error(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	slog.Default().DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	slog.Default().InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	slog.Default().WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	slog.Default().ErrorContext(ctx, msg, args...)
}
