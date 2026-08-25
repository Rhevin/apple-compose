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
// The file is persisted under ~/.apple-compose/hosts so `container start` still
// finds the mount source after stop (container CLI 1.3.0+ rejects missing paths).
func writeProjectHostsFile(project string, svc types.ServiceConfig) (string, error) {
	entries, err := projectHostsEntries(project, svc)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", nil
	}

	dir := namedHostsDir(project, svc.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "hosts")
	var b strings.Builder
	fmt.Fprintln(&b, "127.0.0.1\tlocalhost")
	fmt.Fprintln(&b, "::1\tip6-localhost ip6-loopback")
	for _, e := range entries {
		fmt.Fprintf(&b, "%s\t%s\n", e.ip, strings.Join(e.names, " "))
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func namedHostsDir(project, service string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".apple-compose", "hosts", project, service)
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
	return injectBeforeImage(args, "--volume", hostsPath+":/etc/hosts:ro")
}

// injectBeforeImage inserts a flag/value pair immediately before the image token
// so Apple's CLI does not treat it as the container command.
func injectBeforeImage(args []string, flag, value string) []string {
	insert := []string{flag, value}
	i := indexOfImageArg(args)
	if i < 0 {
		return append(append([]string{}, args...), insert...)
	}
	out := make([]string, 0, len(args)+2)
	out = append(out, args[:i]...)
	out = append(out, insert...)
	out = append(out, args[i:]...)
	return out
}

// indexOfImageArg finds the image token in `container run` args, skipping flag values
// such as --name <id> so we do not insert mounts between a flag and its argument.
func indexOfImageArg(args []string) int {
	i := 0
	if i < len(args) && args[i] == "run" {
		i++
	}
	for i < len(args) {
		a := args[i]
		if a == "--" {
			if i+1 < len(args) {
				return i + 1
			}
			return -1
		}
		if !strings.HasPrefix(a, "-") {
			return i
		}
		if strings.Contains(a, "=") || runFlagTakesNoValue(a) {
			i++
			continue
		}
		i += 2
	}
	return -1
}

func runFlagTakesNoValue(flag string) bool {
	switch flag {
	case "-d", "--detach", "-i", "--interactive", "-t", "--tty",
		"--read-only", "--init", "--rm", "--remove", "--rosetta",
		"--virtualization", "--no-dns", "--ssh":
		return true
	default:
		return false
	}
}
