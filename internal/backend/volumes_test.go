package backend

import (
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func TestRunArgs_NamedVolume(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "db",
		Image: "postgres:16",
		Volumes: []types.ServiceVolumeConfig{
			{Type: "volume", Source: "pgdata", Target: "/var/lib/postgresql/data"},
		},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}

	// Find the --volume flag value
	var volArg string
	for i, a := range args {
		if a == "--volume" && i+1 < len(args) {
			volArg = args[i+1]
			break
		}
	}
	if volArg == "" {
		t.Fatal("expected --volume flag, not found")
	}
	if !strings.Contains(volArg, ".apple-compose/volumes/myapp/pgdata") {
		t.Errorf("expected volume path under .apple-compose/volumes/myapp/pgdata, got %s", volArg)
	}
	if !strings.HasSuffix(volArg, ":/var/lib/postgresql/data") {
		t.Errorf("expected target :/var/lib/postgresql/data, got %s", volArg)
	}
}

func TestRunArgs_NoVolumeSource(t *testing.T) {
	// Anonymous volume (no source) — should be skipped, no --volume flag
	svc := types.ServiceConfig{
		Name:  "app",
		Image: "myapp:latest",
		Volumes: []types.ServiceVolumeConfig{
			{Type: "volume", Source: "", Target: "/tmp/cache"},
		},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	for i, a := range args {
		if a == "--volume" {
			t.Errorf("unexpected --volume flag at index %d with value %s", i, args[i+1])
		}
	}
}
