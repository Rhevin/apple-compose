package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:                "exec <service> <command> [args...]",
	Short:              "Execute a command in a running container",
	Args:               cobra.MinimumNArgs(2),
	DisableFlagParsing: true, // pass all flags through to the container command
	RunE: func(cmd *cobra.Command, args []string) error {
		// With DisableFlagParsing, we must handle -t/-i manually
		var tty, interactive bool
		remaining := []string{}
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "-t", "--tty":
				tty = true
			case "-i", "--interactive":
				interactive = true
			case "-ti", "-it":
				tty = true
				interactive = true
			default:
				remaining = append(remaining, args[i:]...)
				i = len(args) // break
			}
		}

		if len(remaining) < 2 {
			return fmt.Errorf("usage: exec <service> <command> [args...]")
		}

		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		svcName := remaining[0]
		if _, err := project.GetService(svcName); err != nil {
			return serviceNotFound(svcName, project)
		}

		containerName := backend.ContainerName(project.Name, svcName)
		execArgs := []string{"exec"}
		if tty {
			execArgs = append(execArgs, "--tty")
		}
		if interactive {
			execArgs = append(execArgs, "--interactive")
		}
		execArgs = append(execArgs, containerName)
		execArgs = append(execArgs, remaining[1:]...)

		c := exec.Command("container", execArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
