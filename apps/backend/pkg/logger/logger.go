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

// Init initializes the global logger with file rotation and caller info
func Init(logDir string) {
	// Ensure log directory exists
	os.MkdirAll(logDir, os.ModePerm)

	// Nama file aktif tetap acis.log
	logFile := filepath.Join(logDir, "acis.log")

	// Setup Lumberjack for daily log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // Megabytes
		MaxBackups: 30, // Keep 30 backups
		MaxAge:     30, // Days
		Compress:   true, // Compress old logs to .gz
	}

	// Create a multi-writer: write to both file and console
	multiWriter := io.MultiWriter(lumberjackLogger, os.Stdout)

	// Ubah format waktu jadi Human-Readable (ISO 8601)
	zerolog.TimeFieldFormat = time.RFC3339
	
	// Setup Caller info
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := filepath.Base(file)
		return short + ":" + strconv.Itoa(line)
	}

	log.Logger = zerolog.New(multiWriter).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Info().Msg("Logger initialized successfully with daily rotation")
}