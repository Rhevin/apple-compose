package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var (
	runTTY         bool
	runInteractive bool
	runRM          bool
	runEnv         []string
	runEntrypoint  string
)

var runCmd = &cobra.Command{
	Use:   "run <service> [command] [args...]",
	Short: "Run a one-off command on a service",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		svcName := args[0]
		svc, err := project.GetService(svcName)
		if err != nil {
			return serviceNotFound(svcName, project)
		}

		// Build run args from service config but override command
		runArgs, err := backend.RunArgs(project.Name, svc)
		if err != nil {
			return err
		}

		// Rebuild as a foreground one-off: drop --detach, add --rm if requested
		filtered := []string{"run"}
		if runRM {
			filtered = append(filtered, "--rm")
		}
		if runTTY {
			filtered = append(filtered, "--tty")
		}
		if runInteractive {
			filtered = append(filtered, "--interactive")
		}
		if runEntrypoint != "" {
			filtered = append(filtered, "--entrypoint", runEntrypoint)
		}
		for _, e := range runEnv {
			filtered = append(filtered, "--env", e)
		}

		// Copy flags from RunArgs except: run, --detach, --name
		skip := false
		for i := 1; i < len(runArgs); i++ {
			if runArgs[i] == "--detach" {
				continue
			}
			if runArgs[i] == "--name" {
				skip = true
				continue
			}
			if skip {
				skip = false
				continue
			}
			filtered = append(filtered, runArgs[i])
		}

		// Override command if provided
		if len(args) > 1 {
			// strip existing image+command from filtered, then re-append image + new command
			// image is always the last non-command arg; find it via svc.Image
			imgIdx := -1
			for i, a := range filtered {
				if a == svc.Image {
					imgIdx = i
					break
				}
			}
			if imgIdx >= 0 {
				filtered = append(filtered[:imgIdx+1], args[1:]...)
			}
		}

		fmt.Printf("  [run] %s\n", svcName)
		c := exec.Command("container", filtered...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	runCmd.Flags().BoolVarP(&runTTY, "tty", "t", false, "Allocate a TTY")
	runCmd.Flags().BoolVarP(&runInteractive, "interactive", "i", false, "Keep STDIN open")
	runCmd.Flags().BoolVar(&runRM, "rm", true, "Remove container after run")
	runCmd.Flags().StringArrayVarP(&runEnv, "env", "e", nil, "Set environment variables")
	runCmd.Flags().StringVar(&runEntrypoint, "entrypoint", "", "Override the entrypoint")
}
