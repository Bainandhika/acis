package logger

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommaWriter(t *testing.T) {
	var buf bytes.Buffer
	cw := NewCommaWriter(&buf)

	input := "{" + `"level":"info","message":"test"` + "}\n"
	n, err := cw.Write([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(input) {
		t.Fatalf("expected written count %d, got %d", len(input), n)
	}

	output := buf.String()
	expected := "{" + `"level":"info","message":"test"` + "},\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
	if !strings.HasSuffix(output, ",\n") {
		t.Errorf("expected output to end with comma and newline, got %q", output)
	}
}

func TestMonthlyRotator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "acis_log_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockTime := time.Date(2026, time.July, 15, 10, 0, 0, 0, time.UTC)
	timeNow := func() time.Time {
		return mockTime
	}

	rotator, err := NewMonthlyRotatorWithTime(tempDir, "acis.log", timeNow)
	if err != nil {
		t.Fatalf("failed to create monthly rotator: %v", err)
	}

	_, err = rotator.Write([]byte(`{"msg":"july_log"}` + "\n"))
	if err != nil {
		t.Fatalf("failed to write to rotator: %v", err)
	}

	// Advance time to August (calendar month rollover)
	mockTime = time.Date(2026, time.August, 1, 0, 1, 0, 0, time.UTC)

	_, err = rotator.Write([]byte(`{"msg":"august_log"}` + "\n"))
	if err != nil {
		t.Fatalf("failed to write to rotator after month rollover: %v", err)
	}
	_ = rotator.Close()

	// Verify archived file with YYYYMM format: acis-202607.log
	archivedPath := filepath.Join(tempDir, "acis-202607.log")
	archivedContent, err := os.ReadFile(archivedPath)
	if err != nil {
		t.Fatalf("expected archived file %s to exist: %v", archivedPath, err)
	}
	if !strings.Contains(string(archivedContent), "july_log") {
		t.Errorf("expected archived file to contain july_log, got %s", string(archivedContent))
	}

	// Verify active file contains august_log
	activePath := filepath.Join(tempDir, "acis.log")
	activeContent, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("expected active log file %s to exist: %v", activePath, err)
	}
	if !strings.Contains(string(activeContent), "august_log") {
		t.Errorf("expected active file to contain august_log, got %s", string(activeContent))
	}
}

func TestSlogWithCommaWriter(t *testing.T) {
	var buf bytes.Buffer
	cw := NewCommaWriter(&buf)

	handler := slog.NewJSONHandler(cw, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	testLogger := slog.New(handler)

	testLogger.Info("test message", slog.String("key", "value"))

	output := buf.String()
	if !strings.HasSuffix(output, ",\n") {
		t.Errorf("expected output to end with comma and newline, got %q", output)
	}
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Errorf("expected output to contain json message, got %q", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("expected output to contain json key value, got %q", output)
	}
}
