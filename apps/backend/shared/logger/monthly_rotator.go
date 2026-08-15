package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TimeProvider allows mocking current time for testing rotation
type TimeProvider func() time.Time

// MonthlyRotator handles writing to a log file and rotating/archiving every calendar month
type MonthlyRotator struct {
	dir          string
	filename     string
	mu           sync.Mutex
	file         *os.File
	currentYear  int
	currentMonth time.Month
	timeNow      TimeProvider
}

// NewMonthlyRotator creates a new monthly rotator for the given directory and filename
func NewMonthlyRotator(dir string, filename string) (*MonthlyRotator, error) {
	return NewMonthlyRotatorWithTime(dir, filename, time.Now)
}

// NewMonthlyRotatorWithTime creates a new monthly rotator with custom time provider
func NewMonthlyRotatorWithTime(dir string, filename string, timeNow TimeProvider) (*MonthlyRotator, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	if timeNow == nil {
		timeNow = time.Now
	}

	rotator := &MonthlyRotator{
		dir:      dir,
		filename: filename,
		timeNow:  timeNow,
	}

	if err := rotator.openCurrentFile(); err != nil {
		return nil, err
	}

	return rotator, nil
}

func (r *MonthlyRotator) openCurrentFile() error {
	now := r.timeNow()
	targetPath := filepath.Join(r.dir, r.filename)

	// Check if existing file belongs to a previous month
	if info, err := os.Stat(targetPath); err == nil {
		modTime := info.ModTime()
		if modTime.Year() != now.Year() || modTime.Month() != now.Month() {
			// Archive previous month's file with YYYYMM format
			archiveName := fmt.Sprintf("acis-%04d%02d.log", modTime.Year(), modTime.Month())
			archivePath := filepath.Join(r.dir, archiveName)
			_ = os.Rename(targetPath, archivePath)
		}
	}

	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", targetPath, err)
	}

	r.file = file
	r.currentYear = now.Year()
	r.currentMonth = now.Month()
	return nil
}

// Write writes log bytes to the active log file, rotating if calendar month rolled over
func (r *MonthlyRotator) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.timeNow()
	if now.Year() != r.currentYear || now.Month() != r.currentMonth {
		if err := r.rotate(now); err != nil {
			return 0, err
		}
	}

	if r.file == nil {
		if err := r.openCurrentFile(); err != nil {
			return 0, err
		}
	}

	return r.file.Write(p)
}

func (r *MonthlyRotator) rotate(now time.Time) error {
	if r.file != nil {
		_ = r.file.Sync()
		_ = r.file.Close()
		r.file = nil
	}

	targetPath := filepath.Join(r.dir, r.filename)
	archiveName := fmt.Sprintf("acis-%04d%02d.log", r.currentYear, r.currentMonth)
	archivePath := filepath.Join(r.dir, archiveName)

	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Rename(targetPath, archivePath); err != nil {
			return fmt.Errorf("failed to archive log file from %s to %s: %w", targetPath, archivePath, err)
		}
	}

	return r.openCurrentFile()
}

// Close closes the underlying open file
func (r *MonthlyRotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file != nil {
		err := r.file.Close()
		r.file = nil
		return err
	}
	return nil
}

// Sync flushes file buffers to storage
func (r *MonthlyRotator) Sync() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file != nil {
		return r.file.Sync()
	}
	return nil
}
