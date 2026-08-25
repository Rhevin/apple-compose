package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var pullScheme string

var pullCmd = &cobra.Command{
	Use:   "pull [service...]",
	Short: "Pull service images",
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}

		services := args
		if len(services) == 0 {
			for name := range project.Services {
				services = append(services, name)
			}
		}

		for _, name := range services {
			svc, err := project.GetService(name)
			if err != nil {
				return serviceNotFound(name, project)
			}
			if svc.Build != nil && svc.Image == "" {
				fmt.Fprintf(os.Stderr, "  [skip] %s: build-only service, no image to pull\n", name)
				continue
			}
			fmt.Printf("  [pull] %s (%s)\n", name, svc.Image)
			pullArgs, err := backend.ImagePullArgs(svc.Image, svc.Platform, pullScheme)
			if err != nil {
				return err
			}
			c := exec.Command("container", pullArgs...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("pulling %q: %w", svc.Image, err)
			}
		}
		return nil
	},
}

func init() {
	pullCmd.Flags().StringVar(&pullScheme, "scheme", "", "Registry scheme (http or https; default is the container CLI default, https as of 1.3.0)")
}
