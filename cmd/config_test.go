package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withComposeDir(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
}

func captureStdout(fn func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), err
}

func TestConfig_FormatJSON(t *testing.T) {
	withComposeDir(t, `services:
  web:
    image: nginx:alpine
`)
	configFormat = "json"
	configQuiet = false
	configServices = false
	t.Cleanup(func() {
		configFormat = ""
		configQuiet = false
		configServices = false
	})

	out, err := captureStdout(func() error { return configCmd.RunE(configCmd, nil) })
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid json output: %v\n%s", err, out)
	}
	if !strings.Contains(out, "nginx:alpine") {
		t.Fatalf("expected image in output, got: %s", out)
	}
}

func TestConfig_QuietProducesNoOutput(t *testing.T) {
	withComposeDir(t, `services:
  web:
    image: nginx:alpine
    platform: linux/amd64
`)
	configQuiet = true
	configFormat = ""
	configServices = false
	t.Cleanup(func() {
		configQuiet = false
	})

	out, err := captureStdout(func() error { return configCmd.RunE(configCmd, nil) })
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected no stdout with --quiet, got: %q", out)
	}
}
