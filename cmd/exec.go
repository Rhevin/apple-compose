package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var (
	execTTY         bool
	execInteractive bool
)

var execCmd = &cobra.Command{
	Use:   "exec <service> <command> [args...]",
	Short: "Execute a command in a running container",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		svcName := args[0]
		if _, err := project.GetService(svcName); err != nil {
			return serviceNotFound(svcName, project)
		}

		containerName := backend.ContainerName(project.Name, svcName)
		execArgs := []string{"exec"}
		if execTTY {
			execArgs = append(execArgs, "--tty")
		}
		if execInteractive {
			execArgs = append(execArgs, "--interactive")
		}
		execArgs = append(execArgs, containerName)
		execArgs = append(execArgs, args[1:]...)

		c := exec.Command("container", execArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	execCmd.Flags().BoolVarP(&execTTY, "tty", "t", false, "Allocate a TTY")
	execCmd.Flags().BoolVarP(&execInteractive, "interactive", "i", false, "Keep STDIN open")
}
