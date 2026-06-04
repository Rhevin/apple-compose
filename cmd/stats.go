package cmd

import (
	"os"
	"os/exec"

	"github.com/Rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var statsNoStream bool

var statsCmd = &cobra.Command{
	Use:   "stats [service...]",
	Short: "Display resource usage statistics for service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)

		cArgs := []string{"stats"}
		if statsNoStream {
			cArgs = append(cArgs, "--no-stream")
		}
		for _, name := range targets {
			cArgs = append(cArgs, backend.ContainerName(proj, name))
		}

		c := exec.Command("container", cArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	statsCmd.Flags().BoolVar(&statsNoStream, "no-stream", false, "Print a single snapshot instead of streaming")
}
