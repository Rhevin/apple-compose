//go:build integration

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

var (
	binaryPath  string
	composeFile string
)

// TestMain builds the binary once, runs all tests, then cleans up.
func TestMain(m *testing.M) {
	// Must be on macOS arm64
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		os.Exit(0)
	}

	// Must have container CLI available
	if _, err := exec.LookPath("container"); err != nil {
		os.Stderr.WriteString("SKIP: 'container' CLI not found — skipping integration tests\n")
		os.Exit(0)
	}

	// Build the binary
	root := projectRoot()
	bin := filepath.Join(root, "apple-compose-test")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		os.Stderr.WriteString("FATAL: could not build binary: " + string(out))
		os.Exit(1)
	}
	binaryPath = bin
	composeFile = filepath.Join(root, "testdata", "docker-compose.yml")

	code := m.Run()

	// Always remove the test binary
	os.Remove(bin)
	os.Exit(code)
}

// ac runs apple-compose with the given args and returns stdout, stderr, and error.
func ac(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	base := []string{"-f", composeFile}
	cmd := exec.Command(binaryPath, append(base, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// acMust runs apple-compose and fails the test if the command errors.
func acMust(t *testing.T, args ...string) string {
	t.Helper()
	stdout, stderr, err := ac(t, args...)
	if err != nil {
		t.Fatalf("apple-compose %v failed: %v\nstdout: %s\nstderr: %s",
			args, err, stdout, stderr)
	}
	return stdout
}

// cleanup tears down the project regardless of test state.
func cleanup(t *testing.T) {
	t.Helper()
	ac(t, "down") // ignore errors — may already be down
}

// waitFor polls fn every 500ms until it returns true or timeout elapses.
func waitFor(t *testing.T, timeout time.Duration, fn func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

// projectRoot returns the repo root (one level up from integration/).
func projectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(file))
}

// acNoFile runs apple-compose without passing -f (for commands that don't need a compose file).
func acNoFile(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// contains checks if s contains substr (case-sensitive).
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
