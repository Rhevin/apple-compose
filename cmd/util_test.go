package cmd

import (
	"testing"
)

func TestServiceTargets_Empty(t *testing.T) {
	order := []string{"db", "redis", "web"}
	got := serviceTargets(order, nil)
	if len(got) != 3 {
		t.Errorf("expected all 3 services, got %d", len(got))
	}
	for i, name := range order {
		if got[i] != name {
			t.Errorf("pos %d: expected %s, got %s", i, name, got[i])
		}
	}
}

func TestServiceTargets_Subset(t *testing.T) {
	order := []string{"db", "redis", "web"}
	got := serviceTargets(order, []string{"web", "db"})
	if len(got) != 2 {
		t.Fatalf("expected 2 services, got %d", len(got))
	}
	// Order must follow topological order, not the input order
	if got[0] != "db" || got[1] != "web" {
		t.Errorf("expected [db web], got %v", got)
	}
}

func TestServiceTargets_NotFound(t *testing.T) {
	order := []string{"db", "web"}
	got := serviceTargets(order, []string{"cache"})
	if len(got) != 0 {
		t.Errorf("expected empty result for unknown service, got %v", got)
	}
}

func TestServiceTargets_Single(t *testing.T) {
	order := []string{"db", "redis", "web"}
	got := serviceTargets(order, []string{"redis"})
	if len(got) != 1 || got[0] != "redis" {
		t.Errorf("expected [redis], got %v", got)
	}
}

func TestNeedsQuote(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"plain", false},
		{"with space", true},
		{"key=value", true},
		{`has"quote`, true},
		{"normal-flag", false},
	}
	for _, c := range cases {
		if got := needsQuote(c.input); got != c.want {
			t.Errorf("needsQuote(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
