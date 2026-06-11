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
	State string
}

func (s *containerStatusField) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if b[0] == '"' {
		return json.Unmarshal(b, &s.State)
	}
	var obj struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}
	s.State = obj.State
	return nil
}

// appleContainer matches the JSON shape of `container list --format json`.
type appleContainer struct {
	Status        containerStatusField `json:"status"`
	Configuration struct {
		ID    string `json:"id"`
		Image struct {
			Reference string `json:"reference"`
		} `json:"image"`
		Labels map[string]string `json:"labels"`
	} `json:"configuration"`
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

	if svc.MemLimit > 0 {
		args = append(args, "--memory", fmt.Sprintf("%d", int64(svc.MemLimit)))
	}
	if svc.CPUS > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%.2f", svc.CPUS))
	}

	if svc.Restart != "" && svc.Restart != "no" {
		fmt.Fprintf(os.Stderr, "  WARNING: service %q has restart: %q — not supported by Apple container CLI yet, ignored\n", svc.Name, svc.Restart)
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

// Up starts a container for the given service (detached).
// If the container already exists and is running, it is skipped.
// If it exists but is stopped, it is started without recreating.
func Up(project string, svc types.ServiceConfig) error {
	name := ContainerName(project, svc.Name)

	status, err := containerStatus(name)
	if err == nil {
		switch status.State {
		case "running":
			fmt.Printf("  [=] %s (already running)\n", svc.Name)
			return nil
		default:
			fmt.Printf("  [>] %s (restarting stopped container)\n", svc.Name)
			return Start(name)
		}
	}

	// Container doesn't exist — create and start it
	args, err := RunArgs(project, svc)
	if err != nil {
		return err
	}
	fmt.Printf("  [+] %s\n", svc.Name)
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WaitHealthy polls until a container status is "running" or timeout elapses.
func WaitHealthy(project, service string, timeout time.Duration) error {
	name := ContainerName(project, service)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := containerStatus(name)
		if err == nil && status.State == "running" {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("service %q did not become healthy within %s", service, timeout)
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

func containerStatus(name string) (containerStatusField, error) {
	containers, err := listContainers()
	if err != nil {
		return containerStatusField{}, err
	}
	for _, c := range containers {
		if c.Configuration.ID == name {
			return c.Status, nil
		}
	}
	return containerStatusField{}, fmt.Errorf("container %q not found", name)
}

// Stop sends a stop signal to a container without removing it.
func Stop(name string) error {
	return run(bin, "stop", name)
}

// Start starts a previously stopped container.
func Start(name string) error {
	return run(bin, "start", name)
}

// Down stops and removes a container.
func Down(name string) error {
	_ = run(bin, "stop", name)
	return run(bin, "delete", name)
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
	fmt.Fprintf(&buf, "%-30s %-45s %-10s\n", "NAME", "IMAGE", "STATUS")
	fmt.Fprintf(&buf, "%s\n", strings.Repeat("-", 87))
	for _, c := range rows {
		svc := c.Configuration.Labels[LabelService]
		name := ContainerName(project, svc)
		fmt.Fprintf(&buf, "%-30s %-45s %-10s\n", name, c.Configuration.Image.Reference, c.Status.State)
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
