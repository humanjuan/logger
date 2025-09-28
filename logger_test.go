package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper to read whole file into string
func readAll(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer f.Close()
	var b strings.Builder
	s := bufio.NewScanner(f)
	for s.Scan() {
		b.WriteString(s.Text())
		b.WriteByte('\n')
	}
	if err := s.Err(); err != nil {
		t.Fatalf("scan file: %v", err)
	}
	return b.String()
}

func TestLoggerWritesAndFilters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	name := "test.log"

	lg, err := Start(name, tmpDir, Level.INFO)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Use a stable timestamp format (doesn't affect assertions but keeps output predictable)
	lg.TimestampFormat(TS.RFC3339Nano)

	lg.Info("hello")
	lg.Warn("warn")
	lg.Error("err")
	lg.Critical("crit")
	lg.Debug("dbg") // should NOT be logged at INFO level
	lg.Close()

	filePath := filepath.Join(tmpDir, name)
	content := readAll(t, filePath)

	if !strings.HasPrefix(content, "Logger Version:") {
		t.Fatalf("expected header line to start with 'Logger Version:', got: %q", content)
	}

	// Check that expected levels/messages are present
	want := []string{
		"[INFO] hello",
		"[WARN] warn",
		"[ERROR] err",
		"[CRITICAL] crit",
	}
	for _, w := range want {
		if !strings.Contains(content, w) {
			t.Errorf("log does not contain %q", w)
		}
	}

	// Ensure DEBUG message is filtered out
	if strings.Contains(content, "[DEBUG] dbg") {
		t.Errorf("unexpected DEBUG message found when level is INFO")
	}
}

func TestLoggerRotation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger-rotate-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	name := "rotate.log"

	lg, err := Start(name, tmpDir, Level.INFO)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	lg.TimestampFormat(TS.RFC3339)

	// Force rotation on any write by setting max size to 0 MB
	lg.Rotation(0, 2)

	// Use direct Write() to trigger rotation logic (sizeCheck is called only in Write)
	if _, err := lg.Write([]byte("first")); err != nil {
		t.Fatalf("Write error: %v", err)
	}
	lg.Close()

	basePath := filepath.Join(tmpDir, name)
	rot1 := basePath + ".1"

	if _, err := os.Stat(basePath); err != nil {
		t.Fatalf("expected base log file to exist: %v", err)
	}
	if _, err := os.Stat(rot1); err != nil {
		t.Fatalf("expected rotated file .1 to exist: %v", err)
	}

	// Sanity: rotated file should contain the header line
	rotContent := readAll(t, rot1)
	if !strings.HasPrefix(rotContent, "Logger Version:") {
		t.Errorf("rotated file should contain header, got: %q", rotContent)
	}
}
