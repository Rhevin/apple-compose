package backend

import (
	"testing"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
)

func dur(d time.Duration) *types.Duration {
	v := types.Duration(d)
	return &v
}

func TestHealthcheckExecArgs_CMD(t *testing.T) {
	t.Parallel()
	hc := &types.HealthCheckConfig{
		Test: types.HealthCheckTest{"CMD", "curl", "-f", "http://localhost"},
	}
	args, err := HealthcheckExecArgs(hc)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 3 || args[0] != "curl" || args[1] != "-f" || args[2] != "http://localhost" {
		t.Fatalf("got %v", args)
	}
}

func TestHealthcheckExecArgs_CMDShell(t *testing.T) {
	t.Parallel()
	hc := &types.HealthCheckConfig{
		Test: types.HealthCheckTest{"CMD-SHELL", "pg_isready -U app || exit 1"},
	}
	args, err := HealthcheckExecArgs(hc)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 3 || args[0] != "sh" || args[1] != "-c" {
		t.Fatalf("got %v", args)
	}
	if args[2] != "pg_isready -U app || exit 1" {
		t.Fatalf("shell script: got %q", args[2])
	}
}

func TestHealthcheckExecArgs_ImplicitCMD(t *testing.T) {
	t.Parallel()
	hc := &types.HealthCheckConfig{
		Test: types.HealthCheckTest{"redis-cli", "ping"},
	}
	args, err := HealthcheckExecArgs(hc)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != "redis-cli" {
		t.Fatalf("got %v", args)
	}
}

func TestHealthcheckExecArgs_Disabled(t *testing.T) {
	t.Parallel()
	_, err := HealthcheckExecArgs(&types.HealthCheckConfig{Disable: true})
	if err == nil {
		t.Fatal("expected error for disabled healthcheck")
	}
}

func TestHealthcheckInterval_Default(t *testing.T) {
	t.Parallel()
	if got := healthcheckInterval(&types.HealthCheckConfig{}); got != 30*time.Second {
		t.Fatalf("got %s", got)
	}
}

func TestHealthcheckInterval_Custom(t *testing.T) {
	t.Parallel()
	hc := &types.HealthCheckConfig{Interval: dur(5 * time.Second)}
	if got := healthcheckInterval(hc); got != 5*time.Second {
		t.Fatalf("got %s", got)
	}
}
