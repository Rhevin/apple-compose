//go:build integration

package integration

import (
	"strings"
	"testing"
	"time"
)

// TestLifecycle is the main end-to-end test: up → ps → logs → stop → start → down.
// Run with: go test -tags integration -v -timeout 5m ./integration/
func TestLifecycle(t *testing.T) {
	defer cleanup(t)

	// Pull images first so up is faster
	t.Log("pulling images...")
	acMust(t, "pull")

	// Up
	t.Log("starting services...")
	acMust(t, "up", "--wait", "60s")

	// PS — all three services should appear
	t.Log("checking ps...")
	ok := waitFor(t, 30*time.Second, func() bool {
		out := acMust(t, "ps")
		return contains(out, "db") && contains(out, "redis") && contains(out, "web")
	})
	if !ok {
		t.Fatal("ps did not show all services within 30s")
	}

	// Logs — should return output without error
	t.Log("checking logs...")
	acMust(t, "logs", "db")

	// Stop single service
	t.Log("stopping web...")
	acMust(t, "stop", "web")

	// Start it back
	t.Log("starting web...")
	acMust(t, "start", "web")

	// Restart
	t.Log("restarting redis...")
	acMust(t, "restart", "redis")

	// Down
	t.Log("tearing down...")
	acMust(t, "down")

	// PS should show no containers
	out := acMust(t, "ps")
	if contains(out, "myapp-db") || contains(out, "myapp-web") || contains(out, "myapp-redis") {
		t.Errorf("containers still present after down:\n%s", out)
	}
}

func TestUp_DryRun(t *testing.T) {
	out := acMust(t, "up", "--dry-run")
	for _, expect := range []string{"container run", "--detach", "postgres", "redis", "nginx"} {
		if !contains(out, expect) {
			t.Errorf("dry-run output missing %q:\n%s", expect, out)
		}
	}
}

func TestUp_NoDeps(t *testing.T) {
	out := acMust(t, "up", "--dry-run", "--no-deps", "web")
	// Only web should appear, not db or redis
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if contains(line, "container run") {
			if !contains(line, "web") {
				t.Errorf("--no-deps started a service other than web: %s", line)
			}
		}
	}
}
