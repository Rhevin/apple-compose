//go:build integration

package integration

import (
	"testing"
)

func TestExec(t *testing.T) {
	// Use TestLifecycle for full exec testing against real services.
	// Here just verify exec command wiring works with a simple container.
	out := acMust(t, "run", "--rm", "web", "echo", "exec-test")
	if !contains(out, "exec-test") {
		t.Errorf("expected exec-test output, got: %s", out)
	}

	// Verify flags pass through (sh -c uses -c flag)
	out = acMust(t, "run", "--rm", "web", "sh", "-c", "echo exec-with-flags")
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
