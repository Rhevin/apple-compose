package cmd

import (
	"fmt"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var downVolumes bool

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		order, err := topologicalOrder(project)
		if err != nil {
			return err
		}

		fmt.Printf("Stopping project %q\n", project.Name)

		// Tear down in reverse dependency order
		for i := len(order) - 1; i >= 0; i-- {
			name := order[i]
			containerName := backend.ContainerName(project.Name, name)
			fmt.Printf("  [-] %s\n", name)
			svc := project.Services[name]
			opts := backend.StopOptionsFromService(svc)
			if err := backend.Down(containerName, opts); err != nil {
				fmt.Printf("      warning: %v\n", err)
			}
		}

		backend.DeleteNetwork(project.Name)

		if downVolumes {
			fmt.Printf("Removing named volumes for project %q\n", project.Name)
			if err := backend.RemoveNamedVolumes(project.Name); err != nil {
				fmt.Printf("      warning: %v\n", err)
			}
		}
		return nil
	},
}

func init() {
	downCmd.Flags().BoolVarP(&downVolumes, "volumes", "v", false, "Remove named volumes")
}
