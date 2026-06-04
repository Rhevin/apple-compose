package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var topCmd = &cobra.Command{
	Use:   "top [service...]",
	Short: "Display running processes in service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := resolveProjectName()
		targets := resolveTargets(proj, args)
		for _, name := range targets {
			cName := backend.ContainerName(proj, name)
			fmt.Printf("=== %s ===\n", name)
			c := exec.Command("container", "exec", cName, "ps", "aux")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				fmt.Printf("  (could not get processes: %v)\n", err)
			}
			fmt.Println()
		}
		return nil
	},
}
