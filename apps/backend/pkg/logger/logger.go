package logger

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init initializes the global logger with file rotation and caller info
func Init(logDir string) {
	// Ensure log directory exists
	os.MkdirAll(logDir, os.ModePerm)

	// Setup Lumberjack for daily log rotation
	logFile := filepath.Join(logDir, "acis.log")
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,   // Megabytes
		MaxBackups: 30,   // Keep 30 days of logs
		MaxAge:     30,   // Days
		Compress:   true, // Compress old logs
	}

	// Create a multi-writer: write to both file and console (for local dev)
	multiWriter := io.MultiWriter(lumberjackLogger, os.Stdout)

	// Configure Zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		// Show only the filename and line number, not the full path
		short := filepath.Base(file)
		return short + ":" + strconv.Itoa(line)
	}

	log.Logger = zerolog.New(multiWriter).
		With().
		Timestamp().
		Caller(). // Adds file name and line number
		Logger()

	log.Info().Msg("Logger initialized successfully with daily rotation")
}
