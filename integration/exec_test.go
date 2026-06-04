//go:build integration

package integration

import (
	"testing"
)

func TestExec(t *testing.T) {
	// Verify run passes commands through correctly
	out := acMust(t, "run", "--rm", "web", "echo", "exec-test")
	if !contains(out, "exec-test") {
		t.Errorf("expected exec-test output, got: %s", out)
	}
}

func TestRun_OneOff(t *testing.T) {
	out := acMust(t, "run", "--rm", "db", "echo", "one-off")
	if !contains(out, "one-off") {
		t.Errorf("expected one-off output, got: %s", out)
	}
}
