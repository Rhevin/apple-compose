//go:build integration

package integration

import (
	"testing"
	"time"
)

func TestPS_NoComposefile(t *testing.T) {
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
	acMust(t, "up", "--wait", "120s")

	// Wait for containers to appear in ls
	ok := waitFor(t, 60*time.Second, func() bool {
		out, _, err := acNoFile(t, "ls")
		return err == nil && contains(out, "apple-compose")
	})
	if !ok {
		out, _, _ := acNoFile(t, "ls")
		t.Fatalf("expected project 'apple-compose' in ls output, got:\n%s", out)
	}
}
