package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rhevin/apple-compose/internal/compose"
	"github.com/compose-spec/compose-go/v2/types"
)

// topologicalOrder wraps compose.TopologicalOrder for use in cmd package.
func topologicalOrder(project *types.Project) ([]string, error) {
	return compose.TopologicalOrder(project)
}

// serviceNotFound returns a helpful error listing available service names.
func serviceNotFound(name string, project *types.Project) error {
	names := make([]string, 0, len(project.Services))
	for n := range project.Services {
		names = append(names, n)
	}
	sort.Strings(names)
	return fmt.Errorf("service %q not found\n\nAvailable services: %s", name, strings.Join(names, ", "))
}
