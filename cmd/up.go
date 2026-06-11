package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rhevin/apple-compose/internal/backend"
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

			if err := waitForDependencies(project, name, svc); err != nil {
				return err
			}

			if err := backend.Up(project.Name, svc); err != nil {
				return fmt.Errorf("starting service %q: %w", name, err)
			}
		}
		return nil
	},
}

func init() {
	upCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands without executing")
	upCmd.Flags().DurationVar(&healthWait, "wait", 30*time.Second, "Max time to wait for each depends_on condition (0 to disable)")
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

func waitForDependencies(project *types.Project, service string, svc types.ServiceConfig) error {
	if healthWait <= 0 {
		for depName, depCfg := range svc.DependsOn {
			if depCfg.Condition == types.ServiceConditionHealthy {
				fmt.Fprintf(os.Stderr,
					"WARNING: service %q depends on healthy %q but --wait is 0; skipping dependency wait\n",
					service, depName,
				)
			}
		}
		return nil
	}

	for depName, depCfg := range svc.DependsOn {
		depSvc, err := project.GetService(depName)
		if err != nil {
			return fmt.Errorf("service %q: unknown dependency %q: %w", service, depName, err)
		}
		cond := depCfg.Condition
		if cond == "" {
			cond = types.ServiceConditionStarted
		}
		fmt.Printf("      waiting up to %s for %s (%s)...\n", healthWait, depName, cond)
		if err := backend.WaitForDependency(project.Name, depSvc, cond, healthWait); err != nil {
			return fmt.Errorf("dependency %q of %q: %w", depName, service, err)
		}
	}
	return nil
}

func needsQuote(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '=' || c == '"' {
			return true
		}
	}
	return false
}
