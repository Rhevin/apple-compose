package compose

import (
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
)

func TestTopologicalOrder_NoDeps(t *testing.T) {
	project := &types.Project{
		Services: types.Services{
			"a": {Name: "a"},
			"b": {Name: "b"},
		},
	}
	order, err := TopologicalOrder(project)
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 2 {
		t.Fatalf("expected 2 services, got %d", len(order))
	}
}

func TestTopologicalOrder_WithDeps(t *testing.T) {
	project := &types.Project{
		Services: types.Services{
			"web": {
				Name: "web",
				DependsOn: types.DependsOnConfig{
					"db": {Condition: types.ServiceConditionStarted},
				},
			},
			"db": {Name: "db"},
		},
	}
	order, err := TopologicalOrder(project)
	if err != nil {
		t.Fatal(err)
	}
	pos := map[string]int{}
	for i, n := range order {
		pos[n] = i
	}
	if pos["db"] >= pos["web"] {
		t.Errorf("db (pos %d) must come before web (pos %d)", pos["db"], pos["web"])
	}
}

func TestTopologicalOrder_Cycle(t *testing.T) {
	project := &types.Project{
		Services: types.Services{
			"a": {
				Name: "a",
				DependsOn: types.DependsOnConfig{
					"b": {Condition: types.ServiceConditionStarted},
				},
			},
			"b": {
				Name: "b",
				DependsOn: types.DependsOnConfig{
					"a": {Condition: types.ServiceConditionStarted},
				},
			},
		},
	}
	_, err := TopologicalOrder(project)
	if err == nil {
		t.Fatal("expected error for circular dependency, got nil")
	}
}

func TestTopologicalOrder_Chain(t *testing.T) {
	// a → b → c (c must start first)
	project := &types.Project{
		Services: types.Services{
			"a": {
				Name: "a",
				DependsOn: types.DependsOnConfig{
					"b": {Condition: types.ServiceConditionStarted},
				},
			},
			"b": {
				Name: "b",
				DependsOn: types.DependsOnConfig{
					"c": {Condition: types.ServiceConditionStarted},
				},
			},
			"c": {Name: "c"},
		},
	}
	order, err := TopologicalOrder(project)
	if err != nil {
		t.Fatal(err)
	}
	pos := map[string]int{}
	for i, n := range order {
		pos[n] = i
	}
	if pos["c"] >= pos["b"] || pos["b"] >= pos["a"] {
		t.Errorf("expected c < b < a, got c=%d b=%d a=%d", pos["c"], pos["b"], pos["a"])
	}
}
