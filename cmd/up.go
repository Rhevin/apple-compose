package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/Rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var (
	dryRun     bool
	healthWait time.Duration
	noDeps     bool
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create and start containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		// Warn about build: keys before doing anything
		for name, svc := range project.Services {
			if svc.Build != nil {
				fmt.Fprintf(os.Stderr, "WARNING: service %q has a 'build' key — skipped (v0.1 pull-only mode)\n", name)
			}
		}

		order, err := topologicalOrder(project)
		if err != nil {
			return err
		}

		// --no-deps: only start the explicitly named services
		if noDeps && len(args) > 0 {
			order = serviceTargets(order, args)
		}

		if dryRun {
			fmt.Printf("Dry run — project %q (%d services)\n\n", project.Name, len(order))
			for _, name := range order {
				svc := project.Services[name]
				runArgs, err := backend.RunArgs(project.Name, svc)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  [skip] %s: %v\n", name, err)
					continue
				}
				fmt.Printf("  container %s\n", joinArgs(runArgs))
			}
			return nil
		}

		// Create shared network (no-op on macOS < 26)
		if err := backend.EnsureNetwork(project.Name); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not create network: %v\n", err)
		}

		fmt.Printf("Starting project %q (%d services)\n", project.Name, len(order))

		for _, name := range order {
			svc := project.Services[name]
			if svc.Build != nil {
				continue // already warned above
			}

			if err := backend.Up(project.Name, svc); err != nil {
				return fmt.Errorf("starting service %q: %w", name, err)
			}

			// Wait for service to be running before starting dependents
			if healthWait > 0 {
				fmt.Printf("      waiting up to %s for %s...\n", healthWait, name)
				if err := backend.WaitHealthy(project.Name, name, healthWait); err != nil {
					fmt.Fprintf(os.Stderr, "  WARNING: %v\n", err)
				}
			}
		}
		return nil
	},
}

func init() {
	upCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands without executing")
	upCmd.Flags().DurationVar(&healthWait, "wait", 30*time.Second, "Time to wait for each service to be running (0 to disable)")
	upCmd.Flags().BoolVar(&noDeps, "no-deps", false, "Only start named services, skip dependencies")
}

func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		if needsQuote(a) {
			out += fmt.Sprintf("%q", a)
		} else {
			out += a
		}
	}
	return out
}

func needsQuote(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '=' || c == '"' {
			return true
		}
	}
	return false
}
