//go:build integration

package integration

import (
	"testing"
)

func TestPS_NoComposefile(t *testing.T) {
	// ps should work without a compose file when -p is given
	cmd := binaryPath
	_ = cmd
	out, _, err := acNoFile(t, "-p", "nonexistent-project", "ps")
	if err != nil {
		t.Fatalf("ps with -p should not error: %v\nout: %s", err, out)
	}
	if !contains(out, "No containers found") {
		t.Errorf("expected 'No containers found', got: %s", out)
	}
}

func TestLS(t *testing.T) {
	defer cleanup(t)
	acMust(t, "pull")
	acMust(t, "up", "--wait", "60s")

	// ls should show the project without needing -f
	out, _, err := acNoFile(t, "ls")
	if err != nil {
		t.Fatalf("ls failed: %v\nout: %s", err, out)
	}
	if !contains(out, "myapp") {
		t.Errorf("expected project 'myapp' in ls output, got:\n%s", out)
	}
}
