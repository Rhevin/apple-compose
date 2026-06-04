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
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: exec <service> <command> [args...]")
		}

		proj := resolveProjectName()
		svcName := args[0]
		containerName := backend.ContainerName(proj, svcName)

		execArgs := []string{"exec", containerName}
		execArgs = append(execArgs, args[1:]...)

		c := exec.Command("container", execArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	execCmd.Flags().BoolP("tty", "t", false, "Allocate a TTY")
	execCmd.Flags().BoolP("interactive", "i", false, "Keep STDIN open")
}
