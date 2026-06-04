package backend

import (
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func strPtr(s string) *string { return &s }

func TestRunArgs_BasicService(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--name", "myapp-web")
	assertContains(t, args, "--network", "myapp_default")
	assertArg(t, args, "nginx:alpine")
}

func TestRunArgs_PortMapping(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
		Ports: []types.ServicePortConfig{
			{Published: "8080", Target: 80},
		},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--publish", "8080:80")
}

func TestRunArgs_EnvVars(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "db",
		Image: "postgres:16",
		Environment: types.MappingWithEquals{
			"POSTGRES_PASSWORD": strPtr("secret"),
		},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--env", "POSTGRES_PASSWORD=secret")
}

func TestRunArgs_BuildKeyReturnsError(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "app",
		Image: "myapp:latest",
		Build: &types.BuildConfig{Context: "."},
	}
	_, err := RunArgs("myapp", svc)
	if err == nil {
		t.Fatal("expected error for build key, got nil")
	}
	if !strings.Contains(err.Error(), "build") {
		t.Errorf("error should mention 'build', got: %v", err)
	}
}

func TestRunArgs_Labels(t *testing.T) {
	svc := types.ServiceConfig{Name: "web", Image: "nginx:alpine"}
	args, err := RunArgs("proj", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--label", "com.apple-compose.project=proj")
	assertContains(t, args, "--label", "com.apple-compose.service=web")
}

func TestRunArgs_BindMount(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "db",
		Image: "postgres:16",
		Volumes: []types.ServiceVolumeConfig{
			{Type: "bind", Source: "/host/data", Target: "/var/lib/postgresql/data"},
		},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--volume", "/host/data:/var/lib/postgresql/data")
}

func TestRunArgs_RestartWarning(t *testing.T) {
	svc := types.ServiceConfig{
		Name:    "web",
		Image:   "nginx:alpine",
		Restart: "always",
	}
	// Should not error — just warn to stderr
	_, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContainerName(t *testing.T) {
	if got := ContainerName("myapp", "web"); got != "myapp-web" {
		t.Errorf("expected myapp-web, got %s", got)
	}
}

func TestNetworkName(t *testing.T) {
	if got := NetworkName("myapp"); got != "myapp_default" {
		t.Errorf("expected myapp_default, got %s", got)
	}
}

// assertContains checks that key and value appear consecutively in args.
func assertContains(t *testing.T, args []string, key, value string) {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == key && args[i+1] == value {
			return
		}
	}
	t.Errorf("expected %q %q in args: %v", key, value, args)
}

// assertArg checks that a standalone value appears anywhere in args.
func assertArg(t *testing.T, args []string, value string) {
	t.Helper()
	for _, a := range args {
		if a == value {
			return
		}
	}
	t.Errorf("expected %q in args: %v", value, args)
}
