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

	// Wait for postgres to be ready — use pg_isready without flags
	// (avoids cobra intercepting -U as its own flag)
	ok := waitFor(t, 2*time.Minute, func() bool {
		_, _, err := ac(t, "exec", "db", "pg_isready")
		return err == nil
	})
	if !ok {
		t.Fatal("db did not become ready within 2 minutes")
	}

	out := acMust(t, "exec", "db", "echo", "hello-from-container")
	if !contains(out, "hello-from-container") {
		t.Errorf("expected echo output, got: %s", out)
	}
}

func TestRun_OneOff(t *testing.T) {
	out := acMust(t, "run", "--rm", "db", "echo", "one-off")
	if !contains(out, "one-off") {
		t.Errorf("expected one-off output, got: %s", out)
	}
}
