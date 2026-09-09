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
	assertContains(t, args, "--label", "com.apple-compose.config-hash="+serviceConfigHash("myapp", svc))
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

func TestRunArgs_RestartIgnored(t *testing.T) {
	svc := types.ServiceConfig{
		Name:    "web",
		Image:   "nginx:alpine",
		Restart: "always",
	}
	_, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys := UnsupportedServiceKeys(svc)
	if !containsKey(keys, "restart") {
		t.Fatalf("expected restart in unsupported keys, got %v", keys)
	}
}

func TestRunArgs_Platform(t *testing.T) {
	svc := types.ServiceConfig{
		Name:     "web",
		Image:    "nginx:alpine",
		Platform: "linux/amd64",
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--platform", "linux/amd64")
}

func TestRunArgs_DNS(t *testing.T) {
	svc := types.ServiceConfig{
		Name:      "web",
		Image:     "nginx:alpine",
		DNS:       types.StringList{"1.1.1.1", "8.8.8.8"},
		DNSSearch: types.StringList{"example.com"},
		DNSOpts:   types.StringList{"ndots:5"},
	}
	args, err := RunArgs("myapp", svc)
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, args, "--dns", "1.1.1.1")
	assertContains(t, args, "--dns", "8.8.8.8")
	assertContains(t, args, "--dns-search", "example.com")
	assertContains(t, args, "--dns-option", "ndots:5")
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

func TestContainerStatusField_UnmarshalNetworks(t *testing.T) {
	var s containerStatusField
	raw := `{"state":"running","networks":[{"network":"p_default","ipv4Address":"192.168.64.2/24"}]}`
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if len(s.Networks) != 1 || s.Networks[0].IPv4Address != "192.168.64.2/24" {
		t.Fatalf("networks: %+v", s.Networks)
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

// TestAppleContainer_UnmarshalContainerCLI141JSON verifies parsing of real
// container list --format json output from container CLI 1.4.1+, which no longer
// escapes forward slashes in paths and registry references (apple/container#2205).
func TestAppleContainer_UnmarshalContainerCLI141JSON(t *testing.T) {
	raw := `[{
		"id": "testdata-redis",
		"status": {
			"networks": [{
				"hostname": "testdata-redis",
				"ipv4Address": "192.168.66.6/24",
				"network": "testdata_default"
			}],
			"state": "running"
		},
		"configuration": {
			"id": "testdata-redis",
			"image": {"reference": "docker.io/library/redis:8-alpine"},
			"labels": {
				"com.apple-compose.config-hash": "ae883a2cedcf3bc0",
				"com.apple-compose.hosts-hash": "ffeb93371fb7c173",
				"com.apple-compose.project": "testdata",
				"com.apple-compose.service": "redis"
			},
			"publishedPorts": []
		}
	}]`
	var containers []appleContainer
	if err := json.Unmarshal([]byte(raw), &containers); err != nil {
		t.Fatal(err)
	}
	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}
	c := containers[0]
	if c.Configuration.Image.Reference != "docker.io/library/redis:8-alpine" {
		t.Errorf("image reference: got %q", c.Configuration.Image.Reference)
	}
	if c.Status.State != "running" {
		t.Errorf("status: got %q", c.Status.State)
	}
	if len(c.Status.Networks) != 1 || c.Status.Networks[0].IPv4Address != "192.168.66.6/24" {
		t.Fatalf("networks: %+v", c.Status.Networks)
	}
	if c.Configuration.Labels[LabelService] != "redis" {
		t.Errorf("service label: got %q", c.Configuration.Labels[LabelService])
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
