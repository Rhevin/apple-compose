package backend

import (
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func TestServiceConfigHash_Stable(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
		Ports: []types.ServicePortConfig{{Published: "8080", Target: 80}},
		Environment: types.MappingWithEquals{
			"FOO": strPtr("bar"),
		},
	}
	h1 := serviceConfigHash("myapp", svc)
	h2 := serviceConfigHash("myapp", svc)
	if h1 != h2 {
		t.Fatalf("hash not stable: %q vs %q", h1, h2)
	}
}

func TestServiceConfigHash_ChangesOnImage(t *testing.T) {
	base := types.ServiceConfig{Name: "web", Image: "nginx:alpine"}
	changed := types.ServiceConfig{Name: "web", Image: "nginx:latest"}
	if serviceConfigHash("myapp", base) == serviceConfigHash("myapp", changed) {
		t.Fatal("expected different hash for different image")
	}
}

func TestServiceConfigHash_ChangesOnEnv(t *testing.T) {
	a := types.ServiceConfig{
		Name:  "db",
		Image: "postgres:16",
		Environment: types.MappingWithEquals{
			"POSTGRES_PASSWORD": strPtr("a"),
		},
	}
	b := types.ServiceConfig{
		Name:  "db",
		Image: "postgres:16",
		Environment: types.MappingWithEquals{
			"POSTGRES_PASSWORD": strPtr("b"),
		},
	}
	if serviceConfigHash("myapp", a) == serviceConfigHash("myapp", b) {
		t.Fatal("expected different hash for different env")
	}
}

func TestServiceConfigHash_ChangesOnCommand(t *testing.T) {
	a := types.ServiceConfig{Name: "web", Image: "nginx:alpine", Command: []string{"a"}}
	b := types.ServiceConfig{Name: "web", Image: "nginx:alpine", Command: []string{"b"}}
	if serviceConfigHash("myapp", a) == serviceConfigHash("myapp", b) {
		t.Fatal("expected different hash for different command")
	}
}

func TestServiceConfigHash_ChangesOnShmSize(t *testing.T) {
	a := types.ServiceConfig{Name: "web", Image: "nginx:alpine", ShmSize: types.UnitBytes(64 * 1024 * 1024)}
	b := types.ServiceConfig{Name: "web", Image: "nginx:alpine", ShmSize: types.UnitBytes(128 * 1024 * 1024)}
	if serviceConfigHash("myapp", a) == serviceConfigHash("myapp", b) {
		t.Fatal("expected different hash for different shm_size")
	}
}

func TestConfigChanged_HashLabel(t *testing.T) {
	svc := types.ServiceConfig{Name: "web", Image: "nginx:alpine"}
	c := appleContainer{}
	c.Configuration.Labels = map[string]string{
		LabelConfigHash: "stale",
	}
	if !configChanged("myapp", c, svc) {
		t.Fatal("expected drift when hash label differs")
	}
}

func TestConfigChanged_LegacyImage(t *testing.T) {
	svc := types.ServiceConfig{Name: "web", Image: "nginx:alpine"}
	c := appleContainer{}
	c.Configuration.Image.Reference = "nginx:latest"
	if !configChanged("myapp", c, svc) {
		t.Fatal("expected drift when image differs")
	}
}

func TestConfigChanged_LegacyImageDigest(t *testing.T) {
	svc := types.ServiceConfig{Name: "web", Image: "nginx:alpine"}
	c := appleContainer{}
	c.Configuration.Image.Reference = "nginx:alpine@sha256:abc"
	if configChanged("myapp", c, svc) {
		t.Fatal("digest suffix should match desired tag")
	}
}

func TestConfigChanged_LegacyPorts(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
		Ports: []types.ServicePortConfig{{Published: "8080", Target: 80}},
	}
	c := appleContainer{}
	c.Configuration.Image.Reference = "nginx:alpine"
	c.Configuration.PublishedPorts = []publishedPort{{HostPort: 9090, ContainerPort: 80}}
	if !configChanged("myapp", c, svc) {
		t.Fatal("expected drift when ports differ")
	}
}

func TestConfigChanged_NoDrift(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
		Ports: []types.ServicePortConfig{{Published: "8080", Target: 80}},
	}
	c := appleContainer{}
	c.Configuration.Labels = map[string]string{LabelConfigHash: serviceConfigHash("myapp", svc)}
	c.Configuration.Image.Reference = "nginx:alpine"
	c.Configuration.PublishedPorts = []publishedPort{{HostPort: 8080, ContainerPort: 80}}
	if configChanged("myapp", c, svc) {
		t.Fatal("expected no drift when hash matches")
	}
}
