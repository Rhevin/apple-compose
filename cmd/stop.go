package cmd

import (
	"fmt"
	"os/exec"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [service...]",
	Short: "Stop running containers without removing them",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)
		project, _ := loadProject()
		fmt.Printf("Stopping project %q\n", proj)
		for i := len(targets) - 1; i >= 0; i-- {
			name := targets[i]
			fmt.Printf("  [stop] %s\n", name)
			opts := stopOptionsForService(project, name)
			if err := backend.Stop(backend.ContainerName(proj, name), opts); err != nil {
				fmt.Printf("         warning: %v\n", err)
			}
		}
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start [service...]",
	Short: "Start stopped containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)
		fmt.Printf("Starting project %q\n", proj)
		for _, name := range targets {
			fmt.Printf("  [start] %s\n", name)
			if err := backend.Start(backend.ContainerName(proj, name)); err != nil {
				return fmt.Errorf("starting %q: %w", name, err)
			}
		}
		return nil
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart [service...]",
	Short: "Restart service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)
		project, _ := loadProject()
		fmt.Printf("Restarting project %q\n", proj)
		for i := len(targets) - 1; i >= 0; i-- {
			fmt.Printf("  [stop] %s\n", targets[i])
			opts := stopOptionsForService(project, targets[i])
			_ = backend.Stop(backend.ContainerName(proj, targets[i]), opts)
		}
		for _, name := range targets {
			fmt.Printf("  [start] %s\n", name)
			if err := backend.Start(backend.ContainerName(proj, name)); err != nil {
				return fmt.Errorf("starting %q: %w", name, err)
			}
		}
		return nil
	},
}

var killSignal string

var killCmd = &cobra.Command{
	Use:   "kill [service...]",
	Short: "Force stop service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)
		fmt.Printf("Killing project %q\n", proj)
		for i := len(targets) - 1; i >= 0; i-- {
			name := targets[i]
			cName := backend.ContainerName(proj, name)
			fmt.Printf("  [kill] %s\n", name)
			kArgs := []string{"kill"}
			if killSignal != "" {
				kArgs = append(kArgs, "--signal", killSignal)
			}
			kArgs = append(kArgs, cName)
			c := exec.Command("container", kArgs...)
			if out, err := c.CombinedOutput(); err != nil {
				fmt.Printf("         warning: %s\n", string(out))
			}
		}
		return nil
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm [service...]",
	Short: "Remove stopped service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)
		fmt.Printf("Removing containers for project %q\n", proj)
		for i := len(targets) - 1; i >= 0; i-- {
			name := targets[i]
			cName := backend.ContainerName(proj, name)
			fmt.Printf("  [rm] %s\n", name)
			c := exec.Command("container", "delete", cName)
			if out, err := c.CombinedOutput(); err != nil {
				fmt.Printf("       warning: %s\n", string(out))
			}
		}
		return nil
	},
}

// serviceTargets returns the ordered subset of services to act on.
func serviceTargets(order []string, names []string) []string {
	if len(names) == 0 {
		return order
	}
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	var out []string
	for _, n := range order {
		if set[n] {
			out = append(out, n)
		}
	}
	return out
}

func init() {
	killCmd.Flags().StringVarP(&killSignal, "signal", "s", "", "Signal to send (default KILL)")
}
