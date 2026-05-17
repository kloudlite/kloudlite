package registry

import "testing"

func TestDefaultRegistryFindsKnownResources(t *testing.T) {
	r := Default()

	expected := map[string]Scope{
		"users":           Cluster,
		"userpreferences": Cluster,
		"workmachines":    Cluster,
		"machinetypes":    Cluster,
		"workspaces":      Namespaced,
		"environments":    Namespaced,
		"snapshots":       Namespaced,
		"packagerequests": Namespaced,
		"services":        Namespaced,
		"configmaps":      Namespaced,
		"secrets":         Namespaced,
	}

	for alias, scope := range expected {
		resource, ok := r.Get(alias)
		if !ok {
			t.Fatalf("expected %s to be registered", alias)
		}
		if resource.Scope != scope {
			t.Fatalf("expected %s to be %s-scoped, got %s", alias, scope, resource.Scope)
		}
	}

	if len(r.All()) != len(expected) {
		t.Fatalf("expected %d resources, got %d", len(expected), len(r.All()))
	}

	workspace, ok := r.Get("workspaces")
	if !ok {
		t.Fatal("expected workspaces to be registered")
	}
	if workspace.Scope != Namespaced {
		t.Fatalf("expected workspaces to be namespaced, got %s", workspace.Scope)
	}
	if workspace.Kind != "Workspace" || workspace.Plural != "workspaces" {
		t.Fatalf("unexpected workspace metadata: %#v", workspace)
	}

	workMachine, ok := r.Get("workmachines")
	if !ok {
		t.Fatal("expected workmachines to be registered")
	}
	if workMachine.Scope != Cluster {
		t.Fatalf("expected workmachines to be cluster-scoped, got %s", workMachine.Scope)
	}
}

func TestRegistryRejectsDuplicateAliases(t *testing.T) {
	_, err := New([]Resource{
		{Alias: "users", Scope: Cluster},
		{Alias: "users", Scope: Cluster},
	})
	if err == nil {
		t.Fatal("expected duplicate alias error")
	}
}

func TestRegistryRejectsBlankAliases(t *testing.T) {
	_, err := New([]Resource{{Scope: Cluster}})
	if err == nil {
		t.Fatal("expected blank alias error")
	}
}

func TestRegistryRejectsInvalidScopes(t *testing.T) {
	_, err := New([]Resource{{Alias: "users", Scope: Scope("invalid")}})
	if err == nil {
		t.Fatal("expected invalid scope error")
	}
}

func TestResourceScopeValidation(t *testing.T) {
	r := Default()

	if err := r.RequireScope("workspaces", Namespaced); err != nil {
		t.Fatalf("expected namespaced workspaces route to be valid: %v", err)
	}
	if err := r.RequireScope("workspaces", Cluster); err == nil {
		t.Fatal("expected cluster route for workspaces to be invalid")
	}
	if err := r.RequireScope("does-not-exist", Cluster); err == nil {
		t.Fatal("expected unknown resource to be invalid")
	}
}
