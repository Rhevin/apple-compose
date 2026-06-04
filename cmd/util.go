package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/rhevin/apple-compose/internal/backend"
	"github.com/rhevin/apple-compose/internal/compose"
)

// topologicalOrder wraps compose.TopologicalOrder for use in cmd package.
func topologicalOrder(project *types.Project) ([]string, error) {
	return compose.TopologicalOrder(project)
}

// resolveContainerName returns the full container name for a given service name or
// container name. If the user passes the full container name (e.g. "testdata-db")
// it is returned as-is. If they pass a service name (e.g. "db") it is prefixed.
func resolveContainerName(proj, nameOrService string) string {
	prefix := proj + "-"
	if strings.HasPrefix(nameOrService, prefix) {
		return nameOrService // already a full container name
	}
	return backend.ContainerName(proj, nameOrService)
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
