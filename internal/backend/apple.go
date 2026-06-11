package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
)

const bin = "container"

// Label keys stamped on every container we start.
const (
	LabelProject = "com.apple-compose.project"
	LabelService = "com.apple-compose.service"
)

// containerStatusField handles container 1.0.0+ status objects and pre-1.0 strings.
type containerStatusField struct {
	State    string
	Networks []containerNetworkStatus
}

func (s *containerStatusField) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if b[0] == '"' {
		return json.Unmarshal(b, &s.State)
	}
	var obj struct {
		State    string                   `json:"state"`
		Networks []containerNetworkStatus `json:"networks"`
	}
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}
	s.State = obj.State
	s.Networks = obj.Networks
	return nil
}

type publishedPort struct {
	ContainerPort int    `json:"containerPort"`
	HostPort      int    `json:"hostPort"`
	HostAddress   string `json:"hostAddress"`
	Proto         string `json:"proto"`
}

// appleContainer matches the JSON shape of `container list --format json`.
type appleContainer struct {
	Status        containerStatusField `json:"status"`
	Configuration struct {
		ID    string `json:"id"`
		Image struct {
			Reference string `json:"reference"`
		} `json:"image"`
		Labels         map[string]string `json:"labels"`
		PublishedPorts []publishedPort   `json:"publishedPorts"`
	} `json:"configuration"`
}

// StopOptions configures how a container is stopped (maps to container stop flags).
type StopOptions struct {
	Signal  string // e.g. SIGTERM
	Timeout int    // seconds before SIGKILL; 0 uses container CLI default
}

// StopOptionsFromService maps compose stop_signal and stop_grace_period.
func StopOptionsFromService(svc types.ServiceConfig) StopOptions {
	opts := StopOptions{}
	if svc.StopSignal != "" {
		opts.Signal = svc.StopSignal
	}
	if svc.StopGracePeriod != nil {
		secs := int(time.Duration(*svc.StopGracePeriod).Seconds())
		if secs > 0 {
			opts.Timeout = secs
		}
	}
	return opts
}

// ContainerName returns the canonical name for a service container.
func ContainerName(project, service string) string {
	return fmt.Sprintf("%s-%s", project, service)
}

// NetworkName returns the shared network name for a project.
func NetworkName(project string) string {
	return fmt.Sprintf("%s_default", project)
}

// EnsureNetwork creates the project network if it doesn't exist.
// Network commands require macOS 26+; on older systems this is a no-op.
func EnsureNetwork(project string) error {
	name := NetworkName(project)
	out, err := exec.Command(bin, "network", "list", "--format", "json").Output()
	if err != nil {
		return nil // network subcommand not available (macOS < 26)
	}

	var networks []struct {
		ID            string `json:"id"`
		Name          string `json:"name"` // pre-1.0
		Configuration struct {
			Name string `json:"name"`
		} `json:"configuration"`
	}
	if err := json.Unmarshal(out, &networks); err != nil {
		return nil
	}
	for _, n := range networks {
		if n.ID == name || n.Name == name || n.Configuration.Name == name {
			return nil
		}
	}

	fmt.Printf("  [net] creating network %q\n", name)
	cmd := exec.Command(bin, "network", "create", name)
	cmd.Stdout = os.Stdout
	// Suppress stderr — "already exists" error is not actionable
	_ = cmd.Run()
	return nil
}

// DeleteNetwork removes the project network.
func DeleteNetwork(project string) {
	_ = run(bin, "network", "delete", NetworkName(project))
}

// RunArgs builds the `container run` argument list for a service.
// Call PrepareService first to apply virtiofs DB workarounds.
func RunArgs(project string, svc types.ServiceConfig) ([]string, error) {
	if svc.Build != nil {
		return nil, fmt.Errorf(
			"service %q has a 'build' key — custom builds not supported in v0.1 (pull-only mode); remove 'build' or provide a pre-built image",
			svc.Name,
		)
	}

	name := ContainerName(project, svc.Name)
	network := NetworkName(project)

	args := []string{
		"run", "--detach",
		"--name", name,
		"--label", fmt.Sprintf("%s=%s", LabelProject, project),
		"--label", fmt.Sprintf("%s=%s", LabelService, svc.Name),
		"--label", fmt.Sprintf("%s=%s", LabelConfigHash, serviceConfigHash(svc)),
		"--network", network,
		"--dns-search", network,
	}

	for _, p := range svc.Ports {
		args = append(args, "--publish", fmt.Sprintf("%s:%d", p.Published, p.Target))
	}

	for k, v := range svc.Environment {
		val := ""
		if v != nil {
			val = *v
		}
		args = append(args, "--env", fmt.Sprintf("%s=%s", k, val))
	}

	for _, ef := range svc.EnvFiles {
		if ef.Path != "" {
			args = append(args, "--env-file", ef.Path)
		}
	}

	if svc.Entrypoint != nil {
		if len(svc.Entrypoint) == 0 {
			args = append(args, "--entrypoint", "")
		} else {
			args = append(args, "--entrypoint", strings.Join(svc.Entrypoint, " "))
		}
	}

	if svc.User != "" {
		args = append(args, "--user", svc.User)
	}

	if svc.WorkingDir != "" {
		args = append(args, "--workdir", svc.WorkingDir)
	}

	for _, cap := range svc.CapAdd {
		args = append(args, "--cap-add", cap)
	}

	for _, cap := range svc.CapDrop {
		args = append(args, "--cap-drop", cap)
	}

	for _, tmpfs := range svc.Tmpfs {
		args = append(args, "--tmpfs", tmpfs)
	}

	if svc.ReadOnly {
		args = append(args, "--read-only")
	}

	for name, limit := range svc.Ulimits {
		if limit == nil {
			continue
		}
		var val string
		if limit.Single != 0 {
			val = fmt.Sprintf("%d", limit.Single)
		} else {
			val = fmt.Sprintf("%d:%d", limit.Soft, limit.Hard)
		}
		args = append(args, "--ulimit", fmt.Sprintf("%s=%s", name, val))
	}

	if svc.Init != nil && *svc.Init {
		args = append(args, "--init")
	}

	for _, v := range svc.Volumes {
		switch v.Type {
		case "bind":
			args = append(args, "--volume", fmt.Sprintf("%s:%s", v.Source, v.Target))
		case "volume", "":
			if v.Source != "" {
				hostPath := namedVolumePath(project, v.Source)
				if err := os.MkdirAll(hostPath, 0o755); err != nil {
					return nil, fmt.Errorf("creating volume dir %q: %w", hostPath, err)
				}
				args = append(args, "--volume", fmt.Sprintf("%s:%s", hostPath, v.Target))
			}
		}
	}

	memLimit := svc.MemLimit
	cpus := svc.CPUS
	if svc.Deploy != nil && svc.Deploy.Resources.Limits != nil {
		limits := svc.Deploy.Resources.Limits
		if memLimit == 0 {
			memLimit = limits.MemoryBytes
		}
		if cpus == 0 && limits.NanoCPUs > 0 {
			cpus = limits.NanoCPUs.Value()
		}
	}
	if memLimit > 0 {
		args = append(args, "--memory", fmt.Sprintf("%d", int64(memLimit)))
	}
	if cpus > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%.2f", cpus))
	}
	if svc.ShmSize > 0 {
		args = append(args, "--shm-size", formatByteSize(int64(svc.ShmSize)))
	}

	args = append(args, svc.Image)

	if len(svc.Command) > 0 {
		args = append(args, svc.Command...)
	}

	return args, nil
}

func namedVolumePath(project, volume string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".apple-compose", "volumes", project, volume)
}

// UpOptions configures how Up handles existing containers.
type UpOptions struct {
	ForceRecreate bool
}

// Up starts a container for the given service (detached).
// If the container already exists and is running, it is skipped unless config
// changed or ForceRecreate is set.
// If it exists but is stopped, it is started without recreating unless recreate is needed.
func Up(project string, svc types.ServiceConfig, opts UpOptions) error {
	svc = PrepareService(svc)
	name := ContainerName(project, svc.Name)

	c, err := findContainer(name)
	if err == nil {
		if opts.ForceRecreate || configChanged(c, svc) {
			reason := "config changed"
			if opts.ForceRecreate {
				reason = "force-recreate"
			}
			fmt.Printf("  [~] %s (recreating: %s)\n", svc.Name, reason)
			if err := Down(name, StopOptionsFromService(svc)); err != nil {
				return err
			}
			return createContainer(project, svc)
		}
		switch c.Status.State {
		case "running":
			fmt.Printf("  [=] %s (already running)\n", svc.Name)
			return nil
		default:
			fmt.Printf("  [>] %s (restarting stopped container)\n", svc.Name)
			return Start(name)
		}
	}

	return createContainer(project, svc)
}

func createContainer(project string, svc types.ServiceConfig) error {
	hostsPath, cleanup, err := writeProjectHostsFile(project, svc)
	if err != nil {
		return fmt.Errorf("preparing hosts file: %w", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	args, err := RunArgs(project, svc)
	if err != nil {
		return err
	}
	if hostsPath != "" {
		args = injectHostsMount(args, hostsPath)
	}
	fmt.Printf("  [+] %s\n", svc.Name)
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemoveOrphans deletes containers belonging to the project whose service is
// no longer defined in the compose file.
func RemoveOrphans(project string, services map[string]types.ServiceConfig) error {
	containers, err := ListContainersForProject(project)
	if err != nil {
		return err
	}
	for _, c := range containers {
		if _, ok := services[c.Service]; ok {
			continue
		}
		fmt.Printf("  [-] %s (orphan)\n", c.Service)
		if err := Down(c.Name, StopOptions{}); err != nil {
			fmt.Printf("      warning: %v\n", err)
		}
	}
	return nil
}

// RemoveNamedVolumes deletes on-disk data for named volumes in a project.
func RemoveNamedVolumes(project string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	volDir := filepath.Join(home, ".apple-compose", "volumes", project)
	if _, err := os.Stat(volDir); os.IsNotExist(err) {
		fmt.Printf("  no volumes found for project %q\n", project)
		return nil
	}
	fmt.Printf("  removing %s\n", volDir)
	return os.RemoveAll(volDir)
}

// WaitHealthy blocks until a service is ready: healthy if it defines a healthcheck,
// otherwise running.
func WaitHealthy(project string, svc types.ServiceConfig, timeout time.Duration) error {
	name := ContainerName(project, svc.Name)
	hc := svc.HealthCheck
	if hc != nil && !hc.Disable && len(hc.Test) > 0 {
		return waitHealthcheck(name, svc.Name, hc, timeout)
	}
	return waitRunning(name, svc.Name, timeout)
}

// ContainerRecord is a parsed entry from `container list --format json`.
type ContainerRecord struct {
	Status  string
	Project string
	Service string
}

// ListAllContainers returns every container with apple-compose labels.
func ListAllContainers() ([]ContainerRecord, error) {
	containers, err := listContainers()
	if err != nil {
		return nil, err
	}
	var result []ContainerRecord
	for _, c := range containers {
		proj := c.Configuration.Labels[LabelProject]
		svc := c.Configuration.Labels[LabelService]
		if proj == "" {
			continue
		}
		result = append(result, ContainerRecord{
			Status:  c.Status.State,
			Project: proj,
			Service: svc,
		})
	}
	return result, nil
}

func listContainers() ([]appleContainer, error) {
	out, err := exec.Command(bin, "list", "--all", "--format", "json").Output()
	if err != nil {
		return nil, err
	}
	var containers []appleContainer
	if err := json.Unmarshal(out, &containers); err != nil {
		return nil, err
	}
	return containers, nil
}

// ContainerStatus returns the status of a container by name.
func ContainerStatus(name string) (string, error) {
	s, err := containerStatus(name)
	if err != nil {
		return "", err
	}
	return s.State, nil
}

// ContainerInfo holds basic info about a running container.
type ContainerInfo struct {
	Name    string
	Service string
	Status  string
	Image   string
}

// ListContainersForProject returns all containers belonging to a project.
func ListContainersForProject(project string) ([]ContainerInfo, error) {
	containers, err := listContainers()
	if err != nil {
		return nil, err
	}
	var result []ContainerInfo
	for _, c := range containers {
		if c.Configuration.Labels[LabelProject] == project {
			result = append(result, ContainerInfo{
				Name:    c.Configuration.ID,
				Service: c.Configuration.Labels[LabelService],
				Status:  c.Status.State,
				Image:   c.Configuration.Image.Reference,
			})
		}
	}
	return result, nil
}

func findContainer(name string) (appleContainer, error) {
	containers, err := listContainers()
	if err != nil {
		return appleContainer{}, err
	}
	for _, c := range containers {
		if c.Configuration.ID == name {
			return c, nil
		}
	}
	return appleContainer{}, fmt.Errorf("container %q not found", name)
}

func containerStatus(name string) (containerStatusField, error) {
	c, err := findContainer(name)
	if err != nil {
		return containerStatusField{}, err
	}
	return c.Status, nil
}

// Stop sends a stop signal to a container without removing it.
func Stop(name string, opts StopOptions) error {
	return run(bin, stopArgs(name, opts)...)
}

// Start starts a previously stopped container.
func Start(name string) error {
	return run(bin, "start", name)
}

// Down stops and removes a container.
func Down(name string, opts StopOptions) error {
	_ = run(bin, stopArgs(name, opts)...)
	return run(bin, "delete", name)
}

func stopArgs(name string, opts StopOptions) []string {
	args := []string{"stop"}
	if opts.Signal != "" {
		args = append(args, "--signal", opts.Signal)
	}
	if opts.Timeout > 0 {
		args = append(args, "--time", fmt.Sprintf("%d", opts.Timeout))
	}
	args = append(args, name)
	return args
}

// formatByteSize formats bytes for container CLI size flags (e.g. 64M, 1G).
func formatByteSize(b int64) string {
	switch {
	case b%(1<<30) == 0 && b >= 1<<30:
		return fmt.Sprintf("%dG", b/(1<<30))
	case b%(1<<20) == 0 && b >= 1<<20:
		return fmt.Sprintf("%dM", b/(1<<20))
	case b%(1<<10) == 0 && b >= 1<<10:
		return fmt.Sprintf("%dK", b/(1<<10))
	default:
		return fmt.Sprintf("%d", b)
	}
}

func formatPublishedPorts(ports []publishedPort) string {
	if len(ports) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		host := p.HostAddress
		if host == "" {
			host = "0.0.0.0"
		}
		proto := p.Proto
		if proto == "" {
			proto = "tcp"
		}
		parts = append(parts, fmt.Sprintf("%s:%d->%d/%s", host, p.HostPort, p.ContainerPort, proto))
	}
	return strings.Join(parts, ", ")
}

// PS lists containers belonging to the given project, formatted as a table.
func PS(project string) error {
	containers, err := listContainers()
	if err != nil {
		return run(bin, "list")
	}

	var rows []appleContainer
	for _, c := range containers {
		if c.Configuration.Labels[LabelProject] == project {
			rows = append(rows, c)
		}
	}

	if len(rows) == 0 {
		fmt.Printf("No containers found for project %q\n", project)
		return nil
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%-28s %-10s %-16s %-35s %-10s %s\n",
		"NAME", "SERVICE", "ADDRESS", "IMAGE", "STATUS", "PORTS")
	fmt.Fprintf(&buf, "%s\n", strings.Repeat("-", 125))
	for _, c := range rows {
		svc := c.Configuration.Labels[LabelService]
		name := ContainerName(project, svc)
		fmt.Fprintf(&buf, "%-28s %-10s %-16s %-35s %-10s %s\n",
			name, svc, formatContainerAddresses(c, project),
			c.Configuration.Image.Reference, c.Status.State,
			formatPublishedPorts(c.Configuration.PublishedPorts))
	}
	fmt.Print(buf.String())
	return nil
}

// Logs tails logs for a named container.
func Logs(name string, follow bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "--follow")
	}
	args = append(args, name)
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
