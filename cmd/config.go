package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

var configServices bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Parse and print the resolved compose file",
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		if configServices {
			for name := range project.Services {
				fmt.Println(name)
			}
			return nil
		}

		// Marshal project to JSON then back to YAML for clean output
		b, err := json.Marshal(project)
		if err != nil {
			return err
		}
		var obj interface{}
		if err := json.Unmarshal(b, &obj); err != nil {
			return err
		}
		out, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}
		fmt.Print(string(out))
		return nil
	},
}

func init() {
	configCmd.Flags().BoolVar(&configServices, "services", false, "Print service names, one per line")
}
