package logger

import (
	"bytes"
	"strings"
	"testing"
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
