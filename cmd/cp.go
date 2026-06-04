package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var cpCmd = &cobra.Command{
	Use:   "cp <service:path> <local-path> | <local-path> <service:path>",
	Short: "Copy files between a service container and the local filesystem",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		src := resolveContainerPath(args[0], project)
		dst := resolveContainerPath(args[1], project)

		c := exec.Command("container", "copy", src, dst)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

// resolveContainerPath replaces "service:path" with "containerName:path".
func resolveContainerPath(arg string, project *types.Project) string {
	parts := strings.SplitN(arg, ":", 2)
	if len(parts) != 2 {
		return arg
	}
	svcName := parts[0]
	if _, err := project.GetService(svcName); err != nil {
		return arg // not a service name — pass through as-is
	}
	return backend.ContainerName(project.Name, svcName) + ":" + parts[1]
}
