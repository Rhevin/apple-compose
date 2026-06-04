//go:build integration

package integration

import (
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	defer cleanup(t)

	acMust(t, "pull", "db")
	acMust(t, "up", "--wait", "120s")

	// Wait for db container to be running
	ok := waitFor(t, 2*time.Minute, func() bool {
		out, _, _ := ac(t, "ps")
		return contains(out, "db") && contains(out, "running")
	})
	if !ok {
		out, _, _ := ac(t, "ps")
		t.Fatalf("db did not reach running state within 2 minutes, ps:\n%s", out)
	}

	// Verify exec works with a simple command
	out := acMust(t, "exec", "db", "echo", "hello-from-container")
	if !contains(out, "hello-from-container") {
		t.Errorf("expected echo output, got: %s", out)
	}

	// Verify exec passes flags through to container command correctly
	out = acMust(t, "exec", "db", "sh", "-c", "echo exec-with-flags")
	if !contains(out, "exec-with-flags") {
		t.Errorf("expected exec-with-flags output, got: %s", out)
	}
}

func TestRun_OneOff(t *testing.T) {
	out := acMust(t, "run", "--rm", "db", "echo", "one-off")
	if !contains(out, "one-off") {
		t.Errorf("expected one-off output, got: %s", out)
	}
}
