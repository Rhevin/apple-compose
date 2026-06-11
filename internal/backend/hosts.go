package backend

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

const LabelHostsHash = "com.apple-compose.hosts-hash"

type containerNetworkStatus struct {
	Network     string `json:"network"`
	Hostname    string `json:"hostname"`
	IPv4Address string `json:"ipv4Address"`
}

// projectNetworkIP returns the container IPv4 on the project network, without CIDR suffix.
func projectNetworkIP(c appleContainer, project string) string {
	netName := NetworkName(project)
	for _, n := range c.Status.Networks {
		if n.Network != netName || n.IPv4Address == "" {
			continue
		}
		return strings.SplitN(n.IPv4Address, "/", 2)[0]
	}
	return ""
}

// formatContainerAddresses returns the project-network IPv4 for ps output.
func formatContainerAddresses(c appleContainer, project string) string {
	ip := projectNetworkIP(c, project)
	if ip == "" {
		return "-"
	}
	return ip
}

type hostsEntry struct {
	ip    string
	names []string
}

// writeProjectHostsFile builds an /etc/hosts bind-mount for peer service discovery.
// Returns the file path and a cleanup func, or ("", nil, nil) when no entries are needed.
func writeProjectHostsFile(project string, svc types.ServiceConfig) (string, func(), error) {
	entries, err := projectHostsEntries(project, svc)
	if err != nil {
		return "", nil, err
	}
	if len(entries) == 0 {
		return "", nil, nil
	}

	dir, err := os.MkdirTemp("", "apple-compose-hosts-*")
	if err != nil {
		return "", nil, err
	}
	path := filepath.Join(dir, "hosts")
	var b strings.Builder
	fmt.Fprintln(&b, "127.0.0.1\tlocalhost")
	fmt.Fprintln(&b, "::1\tip6-localhost ip6-loopback")
	for _, e := range entries {
		fmt.Fprintf(&b, "%s\t%s\n", e.ip, strings.Join(e.names, " "))
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		_ = os.RemoveAll(dir)
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	return path, cleanup, nil
}

func projectHostsEntries(project string, svc types.ServiceConfig) ([]hostsEntry, error) {
	containers, err := listContainers()
	if err != nil {
		return nil, err
	}

	byIP := map[string][]string{}
	add := func(ip, name string) {
		if ip == "" || name == "" {
			return
		}
		byIP[ip] = append(byIP[ip], name)
	}

	for _, c := range containers {
		if c.Configuration.Labels[LabelProject] != project {
			continue
		}
		if c.Status.State != "running" {
			continue
		}
		peerSvc := c.Configuration.Labels[LabelService]
		if peerSvc == "" || peerSvc == svc.Name {
			continue
		}
		ip := projectNetworkIP(c, project)
		add(ip, peerSvc)
		add(ip, ContainerName(project, peerSvc))
	}

	for host, ips := range svc.ExtraHosts {
		for _, ip := range ips {
			add(ip, host)
		}
	}

	if len(byIP) == 0 {
		return nil, nil
	}

	ips := make([]string, 0, len(byIP))
	for ip := range byIP {
		ips = append(ips, ip)
	}
	sort.Strings(ips)

	entries := make([]hostsEntry, 0, len(ips))
	for _, ip := range ips {
		names := uniqueSorted(byIP[ip])
		entries = append(entries, hostsEntry{ip: ip, names: names})
	}
	return entries, nil
}

func uniqueSorted(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// hostsPeerHash fingerprints the peer /etc/hosts entries a service should have.
func hostsPeerHash(project string, svc types.ServiceConfig) (string, error) {
	entries, err := projectHostsEntries(project, svc)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", nil
	}
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		parts = append(parts, e.ip+":"+strings.Join(uniqueSorted(e.names), ","))
	}
	sort.Strings(parts)
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return hex.EncodeToString(sum[:8]), nil
}

func hostsStale(project string, c appleContainer, svc types.ServiceConfig) bool {
	want, err := hostsPeerHash(project, svc)
	if err != nil || want == "" {
		return false
	}
	return c.Configuration.Labels[LabelHostsHash] != want
}

// RefreshPeerHosts recreates running peers whose /etc/hosts is missing the newly started service.
func RefreshPeerHosts(project, startedService string, services map[string]types.ServiceConfig) error {
	containers, err := listContainers()
	if err != nil {
		return err
	}
	for _, c := range containers {
		if c.Configuration.Labels[LabelProject] != project {
			continue
		}
		if c.Status.State != "running" {
			continue
		}
		peerName := c.Configuration.Labels[LabelService]
		if peerName == "" || peerName == startedService {
			continue
		}
		svc, ok := services[peerName]
		if !ok {
			continue
		}
		svc = PrepareService(svc)
		if !hostsStale(project, c, svc) {
			continue
		}
		fmt.Printf("  [~] %s (refreshing /etc/hosts for %s)\n", peerName, startedService)
		if err := recreateContainer(project, svc); err != nil {
			return fmt.Errorf("refreshing hosts for %q: %w", peerName, err)
		}
	}
	return nil
}

func injectHostsMount(args []string, hostsPath string) []string {
	mount := []string{"--volume", hostsPath + ":/etc/hosts:ro"}
	for i, a := range args {
		if strings.HasPrefix(a, "-") || a == "run" || a == "detach" {
			continue
		}
		out := make([]string, 0, len(args)+2)
		out = append(out, args[:i]...)
		out = append(out, mount...)
		out = append(out, args[i:]...)
		return out
	}
	return append(append([]string{}, args...), mount...)
}
