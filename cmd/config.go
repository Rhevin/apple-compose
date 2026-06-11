package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

var (
	configServices bool
	configQuiet    bool
	configFormat   string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Parse and print the resolved compose file",
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		backend.WarnUnsupportedKeys(os.Stderr, project)

		if configQuiet {
			return nil
		}

		if configServices {
			for name := range project.Services {
				fmt.Println(name)
			}
			return nil
		}

		switch configFormat {
		case "", "yaml":
			return printConfigYAML(os.Stdout, project)
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(project)
		default:
			return fmt.Errorf("unknown format %q (supported: yaml, json)", configFormat)
		}
	},
}

func printConfigYAML(w io.Writer, project interface{}) error {
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
	_, err = w.Write(out)
	return err
}

func init() {
	configCmd.Flags().BoolVar(&configServices, "services", false, "Print service names, one per line")
	configCmd.Flags().BoolVarP(&configQuiet, "quiet", "q", false, "Only validate the configuration, don't print anything")
	configCmd.Flags().StringVar(&configFormat, "format", "", "Output format (yaml or json)")
}
