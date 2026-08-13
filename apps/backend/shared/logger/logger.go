package logger

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// CommaWriter wraps an io.Writer and appends a comma to every log line written
type CommaWriter struct {
	w io.Writer
}

func NewCommaWriter(w io.Writer) io.Writer {
	return &CommaWriter{w: w}
}

func (cw *CommaWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	buf := make([]byte, 0, len(p)+1)
	if p[len(p)-1] == '\n' {
		buf = append(buf, p[:len(p)-1]...)
		buf = append(buf, ',', '\n')
	} else {
		buf = append(buf, p...)
		buf = append(buf, ',')
	}

	if _, err := cw.w.Write(buf); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Init initializes the global logger with file rotation, caller info, and trailing comma formatting
func Init(logDir string) {
	// Ensure log directory exists
	os.MkdirAll(logDir, os.ModePerm)

	// Log file path
	logFile := filepath.Join(logDir, "acis.log")

	// Setup Lumberjack for daily log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,   // Megabytes
		MaxBackups: 30,   // Keep 30 backups
		MaxAge:     30,   // Days
		Compress:   true, // Compress old logs to .gz
	}

	// Create a multi-writer: write to both file and console with trailing comma appending
	multiWriter := io.MultiWriter(lumberjackLogger, os.Stdout)
	commaWriter := NewCommaWriter(multiWriter)

	// Use ISO 8601 time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Setup Caller info
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := filepath.Base(file)
		return short + ":" + strconv.Itoa(line)
	}

	log.Logger = zerolog.New(commaWriter).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Info().Msg("Logger initialized successfully with daily rotation")
}

