package backend

import (
	"bytes"
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func TestUnsupportedServiceKeys_HostnameAndPrivileged(t *testing.T) {
	svc := types.ServiceConfig{
		Name:       "web",
		Image:      "nginx:alpine",
		Hostname:   "web.local",
		Privileged: true,
	}

	keys := UnsupportedServiceKeys(svc)
	if !containsKey(keys, "hostname") {
		t.Fatalf("expected hostname in %v", keys)
	}
	if !containsKey(keys, "privileged") {
		t.Fatalf("expected privileged in %v", keys)
	}
}

func TestUnsupportedServiceKeys_SupportedServiceEmpty(t *testing.T) {
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
		Ports: []types.ServicePortConfig{{Published: "8080", Target: 80}},
	}
	keys := UnsupportedServiceKeys(svc)
	if len(keys) != 0 {
		t.Fatalf("expected no unsupported keys, got %v", keys)
	}
}

func TestUnsupportedServiceKeys_DeployReplicas(t *testing.T) {
	replicas := 3
	svc := types.ServiceConfig{
		Name:  "web",
		Image: "nginx:alpine",
		Deploy: &types.DeployConfig{
			Replicas: &replicas,
			Resources: types.Resources{
				Limits: &types.Resource{MemoryBytes: 128 * 1024 * 1024},
			},
		},
	}
	keys := UnsupportedServiceKeys(svc)
	if !containsKey(keys, "deploy.replicas") {
		t.Fatalf("expected deploy.replicas in %v", keys)
	}
	if containsKey(keys, "deploy.resources.limits") {
		t.Fatalf("deploy limits should be supported, got %v", keys)
	}
}

func TestHasCustomNetworks_DefaultOnly(t *testing.T) {
	if hasCustomNetworks(map[string]*types.ServiceNetworkConfig{"default": {}}) {
		t.Fatal("default-only network should not be unsupported")
	}
	if !hasCustomNetworks(map[string]*types.ServiceNetworkConfig{"frontend": {}}) {
		t.Fatal("custom network should be unsupported")
	}
}

func TestWarnUnsupportedKeys(t *testing.T) {
	project := &types.Project{
		Services: map[string]types.ServiceConfig{
			"web": {
				Name:     "web",
				Image:    "nginx:alpine",
				Hostname: "web.local",
			},
		},
	}
	var buf bytes.Buffer
	WarnUnsupportedKeys(&buf, project)
	out := buf.String()
	if !strings.Contains(out, `service "web"`) || !strings.Contains(out, "hostname") {
		t.Fatalf("unexpected warning output: %q", out)
	}
}

func TestUnsupportedServiceKeys_PlatformAndDNSNowSupported(t *testing.T) {
	svc := types.ServiceConfig{
		Name:      "web",
		Image:     "nginx:alpine",
		Platform:  "linux/amd64",
		DNS:       types.StringList{"1.1.1.1"},
		DNSSearch: types.StringList{"example.com"},
		DNSOpts:   types.StringList{"ndots:5"},
	}
	keys := UnsupportedServiceKeys(svc)
	for _, k := range []string{"platform", "dns", "dns_search", "dns_opt"} {
		if containsKey(keys, k) {
			t.Errorf("%q should no longer be unsupported, but found in %v", k, keys)
		}
	}
}

func containsKey(keys []string, want string) bool {
	for _, k := range keys {
		if k == want {
			return true
		}
	}
	return false
}
