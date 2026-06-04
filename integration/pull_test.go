//go:build integration

package integration

import (
	"testing"
)

func TestPull_AllServices(t *testing.T) {
	out := acMust(t, "pull")
	for _, svc := range []string{"db", "redis", "web"} {
		if !contains(out, svc) {
			t.Errorf("expected service %q in pull output, got:\n%s", svc, out)
		}
	}
}

func TestPull_SingleService(t *testing.T) {
	out := acMust(t, "pull", "web")
	if !contains(out, "web") {
		t.Errorf("expected 'web' in pull output, got:\n%s", out)
	}
}

func TestPull_UnknownService(t *testing.T) {
	_, _, err := ac(t, "pull", "doesnotexist")
	if err == nil {
		t.Fatal("expected error for unknown service, got nil")
	}
}
