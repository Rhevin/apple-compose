package backend

import (
	"strings"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func TestProjectNetworkIP(t *testing.T) {
	c := appleContainer{}
	c.Status.Networks = []containerNetworkStatus{{
		Network: "myapp_default", IPv4Address: "192.168.64.5/24",
	}}
	if got := projectNetworkIP(c, "myapp"); got != "192.168.64.5" {
		t.Errorf("got %q", got)
	}
	if got := projectNetworkIP(c, "other"); got != "" {
		t.Errorf("expected empty for wrong project, got %q", got)
	}
}

func TestInjectHostsMount(t *testing.T) {
	args := []string{"run", "--detach", "--name", "p-web", "nginx:alpine"}
	got := injectHostsMount(args, "/tmp/hosts")
	want := "/tmp/hosts:/etc/hosts:ro"
	found := false
	for _, a := range got {
		if a == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("hosts mount not injected: %v", got)
	}
	if got[len(got)-1] != "nginx:alpine" {
		t.Fatalf("image should remain last arg group, got %v", got)
	}
}

func TestProjectHostsEntries_ExtraHosts(t *testing.T) {
	svc := types.ServiceConfig{
		Name: "web",
		ExtraHosts: types.HostsList{
			"api.example.com": {"10.0.0.1"},
		},
	}
	byIP := map[string][]string{}
	for host, ips := range svc.ExtraHosts {
		for _, ip := range ips {
			byIP[ip] = append(byIP[ip], host)
		}
	}
	if len(byIP["10.0.0.1"]) != 1 || byIP["10.0.0.1"][0] != "api.example.com" {
		t.Fatalf("extra_hosts merge failed: %v", byIP)
	}
}

func TestFormatContainerAddresses(t *testing.T) {
	c := appleContainer{}
	c.Status.Networks = []containerNetworkStatus{{
		Network: "proj_default", IPv4Address: "10.1.2.3/24",
	}}
	if got := formatContainerAddresses(c, "proj"); got != "10.1.2.3" {
		t.Errorf("got %q", got)
	}
	c.Status.Networks = nil
	if got := formatContainerAddresses(c, "proj"); got != "-" {
		t.Errorf("got %q want -", got)
	}
}

func TestHostsStale_NoPeers(t *testing.T) {
	svc := types.ServiceConfig{Name: "web"}
	c := appleContainer{}
	c.Configuration.Labels = map[string]string{LabelHostsHash: "stale"}
	if hostsStale("proj", c, svc) {
		t.Fatal("expected not stale when no running peers need hosts entries")
	}
}

func TestHostsFileContent(t *testing.T) {
	entries := []hostsEntry{{ip: "10.0.0.2", names: []string{"db", "proj-db"}}}
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(e.ip + "\t" + strings.Join(e.names, " ") + "\n")
	}
	if !strings.Contains(b.String(), "10.0.0.2\tdb proj-db") {
		t.Fatalf("unexpected hosts content: %q", b.String())
	}
}
