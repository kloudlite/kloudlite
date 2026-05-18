package registry

import (
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/types/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/types/snapshot/v1"
	userv1alpha1 "github.com/kloudlite/kloudlite/types/user/v1alpha1"
	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
)

type Scope string

const (
	Cluster    Scope = "cluster"
	Namespaced Scope = "namespaced"
)

type Resource struct {
	Alias   string
	Group   string
	Version string
	Kind    string
	Plural  string
	Scope   Scope

	NewObject func() runtime.Object
	NewList   func() runtime.Object
}

func (r Resource) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: r.Group, Version: r.Version, Kind: r.Kind}
}

type Registry struct {
	resources []Resource
	byAlias   map[string]Resource
}

func New(resources []Resource) (*Registry, error) {
	byAlias := make(map[string]Resource, len(resources))
	registered := make([]Resource, 0, len(resources))

	for _, resource := range resources {
		if resource.Alias == "" {
			return nil, fmt.Errorf("resource alias is required")
		}
		if resource.Scope != Cluster && resource.Scope != Namespaced {
			return nil, fmt.Errorf("resource %q has invalid scope %q", resource.Alias, resource.Scope)
		}
		if _, ok := byAlias[resource.Alias]; ok {
			return nil, fmt.Errorf("resource alias %q is already registered", resource.Alias)
		}

		byAlias[resource.Alias] = resource
		registered = append(registered, resource)
	}

	return &Registry{resources: registered, byAlias: byAlias}, nil
}

func Default() *Registry {
	registry, err := New([]Resource{
		newResource("users", userv1alpha1.GroupVersion, "User", Cluster, func() runtime.Object { return &userv1alpha1.User{} }, func() runtime.Object { return &userv1alpha1.UserList{} }),
		newResource("userpreferences", userv1alpha1.GroupVersion, "UserPreferences", Cluster, func() runtime.Object { return &userv1alpha1.UserPreferences{} }, func() runtime.Object { return &userv1alpha1.UserPreferencesList{} }),
		newResource("machinetypes", workmachinev1.GroupVersion, "MachineType", Cluster, func() runtime.Object { return &workmachinev1.MachineType{} }, func() runtime.Object { return &workmachinev1.MachineTypeList{} }),
		newResource("workmachines", workmachinev1.GroupVersion, "WorkMachine", Cluster, func() runtime.Object { return &workmachinev1.WorkMachine{} }, func() runtime.Object { return &workmachinev1.WorkMachineList{} }),
		newResource("workspaces", workspacev1.GroupVersion, "Workspace", Namespaced, func() runtime.Object { return &workspacev1.Workspace{} }, func() runtime.Object { return &workspacev1.WorkspaceList{} }),
		newResource("environments", environmentv1.SchemeGroupVersion, "Environment", Namespaced, func() runtime.Object { return &environmentv1.Environment{} }, func() runtime.Object { return &environmentv1.EnvironmentList{} }),
		newResource("snapshots", snapshotv1.SchemeGroupVersion, "Snapshot", Namespaced, func() runtime.Object { return &snapshotv1.Snapshot{} }, func() runtime.Object { return &snapshotv1.SnapshotList{} }),
		newResource("packagerequests", packagesv1.GroupVersion, "PackageRequest", Namespaced, func() runtime.Object { return &packagesv1.PackageRequest{} }, func() runtime.Object { return &packagesv1.PackageRequestList{} }),
		newResource("services", corev1.SchemeGroupVersion, "Service", Namespaced, func() runtime.Object { return &corev1.Service{} }, func() runtime.Object { return &corev1.ServiceList{} }),
		newResource("configmaps", corev1.SchemeGroupVersion, "ConfigMap", Namespaced, func() runtime.Object { return &corev1.ConfigMap{} }, func() runtime.Object { return &corev1.ConfigMapList{} }),
		newResource("secrets", corev1.SchemeGroupVersion, "Secret", Namespaced, func() runtime.Object { return &corev1.Secret{} }, func() runtime.Object { return &corev1.SecretList{} }),
	})
	if err != nil {
		panic(err)
	}
	return registry
}

func (r *Registry) Get(alias string) (Resource, bool) {
	resource, ok := r.byAlias[alias]
	return resource, ok
}

func (r *Registry) RequireScope(alias string, scope Scope) error {
	resource, ok := r.Get(alias)
	if !ok {
		return fmt.Errorf("resource %q is not registered", alias)
	}
	if resource.Scope != scope {
		return fmt.Errorf("resource %q is %s-scoped, not %s-scoped", alias, resource.Scope, scope)
	}
	return nil
}

func (r *Registry) All() []Resource {
	return slices.Clone(r.resources)
}

func newResource(alias string, gv schema.GroupVersion, kind string, scope Scope, newObject, newList func() runtime.Object) Resource {
	return Resource{
		Alias:     alias,
		Group:     gv.Group,
		Version:   gv.Version,
		Kind:      kind,
		Plural:    alias,
		Scope:     scope,
		NewObject: newObject,
		NewList:   newList,
	}
}
