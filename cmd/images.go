package cmd

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var imagesCmd = &cobra.Command{
	Use:   "images [service...]",
	Short: "List images used by service containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := loadProject()
		if err != nil {
			return fmt.Errorf("loading compose file: %w", err)
		}
		order, err := topologicalOrder(project)
		if err != nil {
			return err
		}
		targets := serviceTargets(order, args)

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%-20s %-45s\n", "SERVICE", "IMAGE")
		fmt.Fprintf(&buf, "%s\n", strings.Repeat("-", 66))

		// Preserve topological order for display
		seen := map[string]bool{}
		for _, name := range targets {
			if seen[name] {
				continue
			}
			seen[name] = true
			svc := project.Services[name]
			fmt.Fprintf(&buf, "%-20s %-45s\n", name, svc.Image)
		}

		// Append any remaining services not in targets (shouldn't happen, but safe)
		remaining := make([]string, 0)
		for name := range project.Services {
			if !seen[name] {
				remaining = append(remaining, name)
			}
		}
		sort.Strings(remaining)
		for _, name := range remaining {
			svc := project.Services[name]
			fmt.Fprintf(&buf, "%-20s %-45s\n", name, svc.Image)
		}

		fmt.Print(buf.String())
		return nil
	},
}
