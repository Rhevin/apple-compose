package cmd

import (
	"path/filepath"

	"github.com/Rhevin/apple-compose/internal/backend"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return backend.PS(resolveProjectName())
	},
}

// resolveProjectName returns the project name without requiring a compose file.
// Priority: --project-name flag > compose file name: field > directory name.
func resolveProjectName() string {
	if projectName != "" {
		return projectName
	}
	if project, err := loadProject(); err == nil {
		return project.Name
	}
	abs, _ := filepath.Abs(composeFile)
	return filepath.Base(filepath.Dir(abs))
}

// resolveTargets returns service names to act on without requiring a compose file.
// If args provided, uses them directly.
// If no args, tries to load compose file for ordered list; falls back to nil (caller handles).
func resolveTargets(_ string, args []string) []string {
	if len(args) > 0 {
		return args
	}
	if project, err := loadProject(); err == nil {
		if order, err := topologicalOrder(project); err == nil {
			return order
		}
	}
	return nil
}
