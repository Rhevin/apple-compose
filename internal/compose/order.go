package compose

import (
	"fmt"

	"github.com/compose-spec/compose-go/v2/types"
)

// TopologicalOrder returns service names in dependency order (dependencies first).
func TopologicalOrder(project *types.Project) ([]string, error) {
	visited := map[string]bool{}
	inStack := map[string]bool{}
	var order []string

	var visit func(name string) error
	visit = func(name string) error {
		if inStack[name] {
			return fmt.Errorf("circular dependency detected at service %q", name)
		}
		if visited[name] {
			return nil
		}
		inStack[name] = true
		svc, err := project.GetService(name)
		if err != nil {
			return err
		}
		for dep := range svc.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		inStack[name] = false
		visited[name] = true
		order = append(order, name)
		return nil
	}

	for name := range project.Services {
		if err := visit(name); err != nil {
			return nil, err
		}
	}
	return order, nil
}
