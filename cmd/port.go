package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var portCmd = &cobra.Command{
	Use:   "port <service> <private-port>",
	Short: "Print the public port for a service port binding",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		svcName := args[0]
		privatePort := args[1]

		svc, err := project.GetService(svcName)
		if err != nil {
			return serviceNotFound(svcName, project)
		}

		for _, p := range svc.Ports {
			if fmt.Sprintf("%d", p.Target) == privatePort {
				fmt.Printf("0.0.0.0:%s\n", p.Published)
				return nil
			}
		}
		return fmt.Errorf("no published port found for %s:%s", svcName, privatePort)
	},
}
