package logger

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// read the whole file content (separate helper name to avoid duplication)
func readWhole(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return string(b)
}

func TestConcurrentLoggingNoRace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger-conc-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	name := "conc.log"
	lg, err := Start(name, tmpDir, Level.DEBUG)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	lg.TimestampFormat(TS.RFC3339Nano)

	const n = 200
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			lg.Info("msg-%d", i)
		}()
	}
	wg.Wait()
	lg.Close()

	content := readWhole(t, filepath.Join(tmpDir, name))
	for i := 0; i < n; i++ {
		needle := "[INFO] msg-" + strconv.Itoa(i)
		if !strings.Contains(content, needle) {
			t.Fatalf("missing concurrent info line: %s", needle)
		}
	}
}

func TestStartInvalidPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger-badpath-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	bad := filepath.Join(tmpDir, "no_such_dir")
	if _, err := Start("x.log", bad, Level.INFO); err == nil {
		t.Fatalf("expected error when starting with non-existent path, got nil")
	}
}

func TestTimestampFormatRFC3339Nano(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger-ts-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	name := "ts.log"
	lg, err := Start(name, tmpDir, Level.INFO)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	lg.TimestampFormat(TS.RFC3339Nano)
	lg.Info("ts")
	lg.Close()

	content := readWhole(t, filepath.Join(tmpDir, name))
	// Skip header line; look for the first log line
	lines := strings.Split(content, "\n")
	var logLine string
	for _, ln := range lines {
		if strings.Contains(ln, "[INFO] ts") {
			logLine = ln
			break
		}
	}
	if logLine == "" {
		t.Fatalf("could not find info line in content: %q", content)
	}
	// RFC3339-like pattern: 2020-01-02T15:04:05(.digits)Z07:00
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2}) \[INFO\] ts$`)
	if !re.MatchString(logLine) {
		t.Fatalf("timestamp does not look RFC3339Nano, got: %q", logLine)
	}
}

func TestCloseFlushesAllMessages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger-flush-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	name := "flush.log"
	lg, err := Start(name, tmpDir, Level.INFO)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	lg.TimestampFormat(TS.RFC3339)

	const n = 150
	for i := 0; i < n; i++ {
		lg.Info("flush-%d", i)
	}
	lg.Close()

	content := readWhole(t, filepath.Join(tmpDir, name))
	count := 0
	for i := 0; i < n; i++ {
		if strings.Contains(content, "[INFO] flush-"+strconv.Itoa(i)) {
			count++
		}
	}
	if count != n {
		t.Fatalf("expected %d flushed lines, got %d", n, count)
	}
}

func TestRotationPrunesOldBackups(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger-prune-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	name := "prune.log"
	lg, err := Start(name, tmpDir, Level.INFO)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	lg.TimestampFormat(TS.RFC3339)
	lg.Rotation(0, 1) // keep only one backup (.1)

	// Each Write triggers a rotation before writing due to size=0
	if _, err := lg.Write([]byte("a")); err != nil {
		t.Fatal(err)
	}
	if _, err := lg.Write([]byte("b")); err != nil {
		t.Fatal(err)
	}
	if _, err := lg.Write([]byte("c")); err != nil {
		t.Fatal(err)
	}
	lg.Close()

	base := filepath.Join(tmpDir, name)
	if _, err := os.Stat(base); err != nil {
		t.Fatalf("base log missing: %v", err)
	}
	if _, err := os.Stat(base + ".1"); err != nil {
		t.Fatalf(".1 backup missing: %v", err)
	}
	if _, err := os.Stat(base + ".2"); err == nil {
		t.Fatalf("unexpected .2 backup should have been pruned")
	}
}
