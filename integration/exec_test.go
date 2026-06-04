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

	// Wait for postgres to accept connections
	ok := waitFor(t, 90*time.Second, func() bool {
		out, _, err := ac(t, "exec", "db", "pg_isready", "-U", "app")
		return err == nil && contains(out, "accepting connections")
	})
	if !ok {
		t.Fatal("db did not become ready within 90s")
	}

	out := acMust(t, "exec", "db", "echo", "hello-from-container")
	if !contains(out, "hello-from-container") {
		t.Errorf("expected echo output, got: %s", out)
	}
}

func TestRun_OneOff(t *testing.T) {
	// run creates a fresh container; network is created automatically
	out := acMust(t, "run", "--rm", "db", "echo", "one-off")
	if !contains(out, "one-off") {
		t.Errorf("expected one-off output, got: %s", out)
	}
}
