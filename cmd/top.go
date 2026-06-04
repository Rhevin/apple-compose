package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var topCmd = &cobra.Command{
	Use:   "top [service...]",
	Short: "Display running processes in service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)

		if len(targets) == 0 {
			targets = runningServicesForProject(proj)
		}

		if len(targets) == 0 {
			fmt.Printf("No running containers found for project %q\n", proj)
			return nil
		}

		for _, name := range targets {
			cName := backend.ContainerName(proj, name)

			status, err := backend.ContainerStatus(cName)
			if err != nil || status != "running" {
				continue
			}

			// Print full container name as header — matches docker compose top
			fmt.Println(cName)

			out, err := runPSInContainer(cName)
			if err != nil {
				fmt.Printf("  (could not get processes: %v)\n\n", err)
				continue
			}
			printTopOutput(out)
			fmt.Println()
		}
		return nil
	},
}

func runPSInContainer(cName string) (string, error) {
	// Try POSIX-style ps first (works on alpine/busybox)
	c := exec.Command("container", "exec", cName,
		"ps", "-eo", "user,pid,ppid,pcpu,stime,tty,time,comm")
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf
	if err := c.Run(); err != nil {
		buf.Reset()
		// Fall back to ps aux (GNU/procps)
		c2 := exec.Command("container", "exec", cName, "ps", "aux")
		c2.Stdout = &buf
		c2.Stderr = &buf
		if err2 := c2.Run(); err2 != nil {
			return "", err2
		}
	}
	return buf.String(), nil
}

func printTopOutput(psOutput string) {
	scanner := bufio.NewScanner(strings.NewReader(psOutput))
	header := true
	for scanner.Scan() {
		line := scanner.Text()
		// Filter out the ps process itself
		if strings.Contains(line, " ps ") {
			continue
		}
		if header {
			// Print Docker-style header
			fmt.Printf("%-10s %-6s %-6s %-5s %-8s %-6s %-10s %s\n",
				"UID", "PID", "PPID", "C", "STIME", "TTY", "TIME", "CMD")
			header = false
			continue
		}
		fmt.Println(line)
	}
}

func runningServicesForProject(proj string) []string {
	containers, err := backend.ListContainersForProject(proj)
	if err != nil {
		return nil
	}
	var services []string
	for _, c := range containers {
		if c.Status == "running" {
			services = append(services, c.Service)
		}
	}
	return services
}
