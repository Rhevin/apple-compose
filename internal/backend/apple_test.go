package backend

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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
	assertContains(t, args, "--label", "com.apple-compose.config-hash="+serviceConfigHash(svc))
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

func TestRunArgs_Entrypoint(t *testing.T) {
	svc := types.ServiceConfig{
		Name:       "web",
		Image:      "nginx:alpine",
		Entrypoint: types.ShellCommand{"/docker-entrypoint.sh"},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--entrypoint", "/docker-entrypoint.sh")
}

func TestRunArgs_UserAndWorkdir(t *testing.T) {
	svc := types.ServiceConfig{
		Name:       "web",
		Image:      "nginx:alpine",
		User:       "1000:1000",
		WorkingDir: "/app",
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--user", "1000:1000")
	assertContains(t, args, "--workdir", "/app")
}

func TestRunArgs_Capabilities(t *testing.T) {
	svc := types.ServiceConfig{
		Name:    "web",
		Image:   "nginx:alpine",
		CapAdd:  []string{"NET_ADMIN"},
		CapDrop: []string{"ALL"},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--cap-add", "NET_ADMIN")
	assertContains(t, args, "--cap-drop", "ALL")
}

func TestRunArgs_TmpfsAndReadOnly(t *testing.T) {
	svc := types.ServiceConfig{
		Name:     "web",
		Image:    "nginx:alpine",
		Tmpfs:    types.StringList{"/run"},
		ReadOnly: true,
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--tmpfs", "/run")
	assertArg(t, args, "--read-only")
}

func TestRunArgs_Ulimits(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
		Ulimits: map[string]*types.UlimitsConfig{
			"nproc":  {Single: 65535},
			"nofile": {Soft: 1024, Hard: 2048},
		},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--ulimit", "nproc=65535")
	assertContains(t, args, "--ulimit", "nofile=1024:2048")
}

func TestRunArgs_InitAndEnvFile(t *testing.T) {
	init := true
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
		Init:  &init,
		EnvFiles: []types.EnvFile{
			{Path: ".env.service"},
		},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertArg(t, args, "--init")
	assertContains(t, args, "--env-file", ".env.service")
}

func TestRunArgs_ResourceLimits(t *testing.T) {
	svc := types.ServiceConfig{
		Name:     "web",
		Image:    "nginx:alpine",
		MemLimit: 512 * 1024 * 1024,
		CPUS:     1.5,
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--memory", "536870912")
	assertContains(t, args, "--cpus", "1.50")
}

func TestRunArgs_ShmSize(t *testing.T) {
	svc := types.ServiceConfig{
		Name:    "web",
		Image:   "nginx:alpine",
		ShmSize: types.UnitBytes(64 * 1024 * 1024),
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--shm-size", "64M")
}

func TestRunArgs_StopOptionsFromService(t *testing.T) {
	grace := types.Duration(10 * time.Second)
	svc := types.ServiceConfig{
		Name:            "web",
		StopSignal:      "SIGINT",
		StopGracePeriod: &grace,
	}
	opts := StopOptionsFromService(svc)
	if opts.Signal != "SIGINT" {
		t.Errorf("signal: got %q", opts.Signal)
	}
	if opts.Timeout != 10 {
		t.Errorf("timeout: got %d", opts.Timeout)
	}
}

func TestFormatByteSize(t *testing.T) {
	tests := map[int64]string{
		64 * 1024 * 1024:       "64M",
		1 * 1024 * 1024:        "1M",
		2 * 1024 * 1024:        "2M",
		1 * 1024 * 1024 * 1024: "1G",
	}
	for bytes, want := range tests {
		if got := formatByteSize(bytes); got != want {
			t.Errorf("formatByteSize(%d) = %q, want %q", bytes, got, want)
		}
	}
}

func TestFormatPublishedPorts(t *testing.T) {
	got := formatPublishedPorts([]publishedPort{{
		HostAddress: "0.0.0.0", HostPort: 8080, ContainerPort: 80, Proto: "tcp",
	}})
	if got != "0.0.0.0:8080->80/tcp" {
		t.Errorf("got %q", got)
	}
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

func TestContainerStatusField_UnmarshalString(t *testing.T) {
	var s containerStatusField
	if err := json.Unmarshal([]byte(`"running"`), &s); err != nil {
		t.Fatal(err)
	}
	if s.State != "running" {
		t.Errorf("expected running, got %q", s.State)
	}
}

func TestContainerStatusField_UnmarshalObject(t *testing.T) {
	var s containerStatusField
	if err := json.Unmarshal([]byte(`{"state":"stopped","networks":[]}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.State != "stopped" {
		t.Errorf("expected stopped, got %q", s.State)
	}
}

func TestAppleContainer_UnmarshalV1(t *testing.T) {
	raw := `[{"id":"myapp-web","status":{"state":"running","networks":[]},"configuration":{"id":"myapp-web","image":{"reference":"nginx:alpine"},"labels":{"com.apple-compose.project":"myapp","com.apple-compose.service":"web"}}}]`
	var containers []appleContainer
	if err := json.Unmarshal([]byte(raw), &containers); err != nil {
		t.Fatal(err)
	}
	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}
	c := containers[0]
	if c.Status.State != "running" {
		t.Errorf("status: got %q", c.Status.State)
	}
	if c.Configuration.Labels[LabelProject] != "myapp" {
		t.Errorf("project label: got %q", c.Configuration.Labels[LabelProject])
	}
}

func TestAppleContainer_UnmarshalPreV1(t *testing.T) {
	raw := `[{"status":"running","configuration":{"id":"myapp-web","image":{"reference":"nginx:alpine"},"labels":{"com.apple-compose.project":"myapp","com.apple-compose.service":"web"}}}]`
	var containers []appleContainer
	if err := json.Unmarshal([]byte(raw), &containers); err != nil {
		t.Fatal(err)
	}
	if containers[0].Status.State != "running" {
		t.Errorf("status: got %q", containers[0].Status.State)
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
