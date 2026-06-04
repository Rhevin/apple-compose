package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var topCmd = &cobra.Command{
	Use:   "top [service...]",
	Short: "Display running processes in service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)

		// If no targets resolved from compose file, get all running containers for project
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

			fmt.Printf("=== %s ===\n", name)
			c := exec.Command("container", "exec", cName, "ps", "aux")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				// Fall back to ps -e (busybox/alpine)
				c2 := exec.Command("container", "exec", cName, "ps", "-e")
				c2.Stdout = os.Stdout
				c2.Stderr = os.Stderr
				if err2 := c2.Run(); err2 != nil {
					fmt.Printf("  (could not get processes: %v)\n", err2)
				}
			}
			fmt.Println()
		}
		return nil
	},
}

// runningServicesForProject returns service names of running containers for a project
// by querying the container runtime directly — no compose file needed.
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
