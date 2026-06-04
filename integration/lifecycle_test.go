//go:build integration

package integration

import (
	"strings"
	"testing"
	"time"
)

// TestLifecycle is the main end-to-end test: up → ps → logs → stop → start → down.
func TestLifecycle(t *testing.T) {
	defer cleanup(t)

	t.Log("pulling images...")
	acMust(t, "pull")

	t.Log("starting services...")
	acMust(t, "up", "--wait", "120s")

	// PS — all three services should appear
	t.Log("checking ps...")
	ok := waitFor(t, 60*time.Second, func() bool {
		out, _, _ := ac(t, "ps")
		return contains(out, "db") && contains(out, "redis") && contains(out, "web")
	})
	if !ok {
		out, _, _ := ac(t, "ps")
		t.Fatalf("ps did not show all services within 60s, got:\n%s", out)
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
	out, _, _ := ac(t, "ps")
	if contains(out, "apple-compose-db") || contains(out, "apple-compose-web") || contains(out, "apple-compose-redis") {
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
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if contains(line, "container run") {
			if !contains(line, "web") {
				t.Errorf("--no-deps started a service other than web: %s", line)
			}
		}
	}
}
