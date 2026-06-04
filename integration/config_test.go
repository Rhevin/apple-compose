//go:build integration

package integration

import (
	"testing"
)

func TestConfig_PrintsYAML(t *testing.T) {
	out := acMust(t, "config")
	for _, expect := range []string{"postgres", "redis", "nginx"} {
		if !contains(out, expect) {
			t.Errorf("config output missing %q:\n%s", expect, out)
		}
	}
}

func TestConfig_Services(t *testing.T) {
	out := acMust(t, "config", "--services")
	for _, svc := range []string{"db", "redis", "web"} {
		if !contains(out, svc) {
			t.Errorf("config --services missing %q:\n%s", svc, out)
		}
	}
}

func TestImages(t *testing.T) {
	out := acMust(t, "images")
	for _, img := range []string{"postgres", "redis", "nginx"} {
		if !contains(out, img) {
			t.Errorf("images output missing %q:\n%s", img, out)
		}
	}
}

func TestPort(t *testing.T) {
	out := acMust(t, "port", "web", "80")
	if !contains(out, "8080") {
		t.Errorf("expected port 8080, got: %s", out)
	}
}
