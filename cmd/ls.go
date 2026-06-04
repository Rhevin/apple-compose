package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"encoding/json"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List running apple-compose projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := exec.Command("container", "list", "--all", "--format", "json").Output()
		if err != nil {
			return fmt.Errorf("listing containers: %w", err)
		}

		var containers []struct {
			Status        string `json:"status"`
			Configuration struct {
				ID     string            `json:"id"`
				Labels map[string]string `json:"labels"`
			} `json:"configuration"`
		}
		if err := json.Unmarshal(out, &containers); err != nil {
			return fmt.Errorf("parsing container list: %w", err)
		}

		type projectInfo struct {
			services map[string]string
		}
		projects := map[string]*projectInfo{}

		for _, c := range containers {
			proj := c.Configuration.Labels["com.apple-compose.project"]
			svc := c.Configuration.Labels["com.apple-compose.service"]
			if proj == "" {
				continue
			}
			if projects[proj] == nil {
				projects[proj] = &projectInfo{services: map[string]string{}}
			}
			projects[proj].services[svc] = c.Status
		}

		if len(projects) == 0 {
			fmt.Println("No apple-compose projects found.")
			return nil
		}

		names := make([]string, 0, len(projects))
		for n := range projects {
			names = append(names, n)
		}
		sort.Strings(names)

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%-25s %-10s %s\n", "NAME", "STATUS", "SERVICES")
		fmt.Fprintf(&buf, "%s\n", strings.Repeat("-", 60))
		for _, name := range names {
			p := projects[name]
			svcs := make([]string, 0, len(p.services))
			for s := range p.services {
				svcs = append(svcs, s)
			}
			sort.Strings(svcs)
			allRunning := true
			for _, st := range p.services {
				if st != "running" {
					allRunning = false
				}
			}
			status := "running"
			if !allRunning {
				status = "degraded"
			}
			fmt.Fprintf(&buf, "%-25s %-10s %s\n", name, status, strings.Join(svcs, ", "))
		}
		fmt.Print(buf.String())
		return nil
	},
}
