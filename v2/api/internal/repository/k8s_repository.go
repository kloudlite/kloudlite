package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sRepository implements the Repository interface using Kubernetes API
type K8sRepository[T client.Object, L client.ObjectList] struct {
	client     client.Client
	newObject  func() T
	newList    func() L
	objectType reflect.Type
	listType   reflect.Type
}

// NewK8sRepository creates a new Kubernetes-based repository
func NewK8sRepository[T client.Object, L client.ObjectList](
	k8sClient client.Client,
	newObject func() T,
	newList func() L,
) Repository[T, L] {
	// Get types for reflection
	objType := reflect.TypeOf(newObject()).Elem()
	listType := reflect.TypeOf(newList()).Elem()

	return &K8sRepository[T, L]{
		client:     k8sClient,
		newObject:  newObject,
		newList:    newList,
		objectType: objType,
		listType:   listType,
	}
}

// Create creates a new resource in Kubernetes
func (r *K8sRepository[T, L]) Create(ctx context.Context, obj T) error {
	if err := r.client.Create(ctx, obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("resource already exists: %w", err)
		}
		return fmt.Errorf("failed to create resource: %w", err)
	}
	return nil
}

// Get retrieves a resource by name and namespace
func (r *K8sRepository[T, L]) Get(ctx context.Context, namespace, name string) (T, error) {
	obj := r.newObject()

	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	if err := r.client.Get(ctx, key, obj); err != nil {
		if apierrors.IsNotFound(err) {
			if namespace == "" {
				return obj, fmt.Errorf("resource not found %s: %w", name, err)
			}
			return obj, fmt.Errorf("resource not found %s/%s: %w", namespace, name, err)
		}
		return obj, fmt.Errorf("failed to get resource: %w", err)
	}

	return obj, nil
}

// Update updates an existing resource
func (r *K8sRepository[T, L]) Update(ctx context.Context, obj T) error {
	if err := r.client.Update(ctx, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("resource not found: %w", err)
		}
		if apierrors.IsConflict(err) {
			return fmt.Errorf("resource conflict (version mismatch): %w", err)
		}
		return fmt.Errorf("failed to update resource: %w", err)
	}
	return nil
}

// Delete deletes a resource by name and namespace
func (r *K8sRepository[T, L]) Delete(ctx context.Context, namespace, name string) error {
	obj := r.newObject()
	obj.SetName(name)

	// Only set namespace if it's not empty (for namespaced resources)
	if namespace != "" {
		obj.SetNamespace(namespace)
	}

	if err := r.client.Delete(ctx, obj); err != nil {
		if apierrors.IsNotFound(err) {
			if namespace == "" {
				return fmt.Errorf("resource not found %s: %w", name, err)
			}
			return fmt.Errorf("resource not found %s/%s: %w", namespace, name, err)
		}
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

// List lists resources with optional filters
func (r *K8sRepository[T, L]) List(ctx context.Context, namespace string, opts ...ListOption) (L, error) {
	list := r.newList()

	// Apply options
	options := &ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Build client list options
	listOpts := &client.ListOptions{
		Namespace: namespace,
	}

	if options.LabelSelector != "" {
		if selector, err := metav1.ParseToLabelSelector(options.LabelSelector); err == nil {
			listOpts.LabelSelector, _ = metav1.LabelSelectorAsSelector(selector)
		}
	}

	if options.Limit > 0 {
		listOpts.Limit = options.Limit
	}

	if options.Continue != "" {
		listOpts.Continue = options.Continue
	}

	if err := r.client.List(ctx, list, listOpts); err != nil {
		return list, fmt.Errorf("failed to list resources: %w", err)
	}

	return list, nil
}

// Watch watches for changes to resources
func (r *K8sRepository[T, L]) Watch(ctx context.Context, namespace string, opts ...WatchOption) (<-chan WatchEvent[T], error) {
	// TODO: Implement watch functionality using controller-runtime client
	// For now, return not implemented error
	return nil, fmt.Errorf("watch functionality not yet implemented")
}

// K8sClusterRepository implements ClusterRepository for cluster-scoped resources
type K8sClusterRepository[T client.Object, L client.ObjectList] struct {
	client    client.Client
	newObject func() T
	newList   func() L
}

// NewK8sClusterRepository creates a new cluster-scoped repository
func NewK8sClusterRepository[T client.Object, L client.ObjectList](
	k8sClient client.Client,
	newObject func() T,
	newList func() L,
) ClusterRepository[T, L] {
	return &K8sClusterRepository[T, L]{
		client:    k8sClient,
		newObject: newObject,
		newList:   newList,
	}
}

// Create creates a new cluster-scoped resource
func (r *K8sClusterRepository[T, L]) Create(ctx context.Context, obj T) error {
	if err := r.client.Create(ctx, obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("resource already exists: %w", err)
		}
		return fmt.Errorf("failed to create resource: %w", err)
	}
	return nil
}

// Get retrieves a cluster-scoped resource by name
func (r *K8sClusterRepository[T, L]) Get(ctx context.Context, name string) (T, error) {
	obj := r.newObject()
	key := types.NamespacedName{Name: name}

	if err := r.client.Get(ctx, key, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return obj, fmt.Errorf("resource not found %s: %w", name, err)
		}
		return obj, fmt.Errorf("failed to get resource: %w", err)
	}

	return obj, nil
}

// Update updates a cluster-scoped resource
func (r *K8sClusterRepository[T, L]) Update(ctx context.Context, obj T) error {
	if err := r.client.Update(ctx, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("resource not found: %w", err)
		}
		if apierrors.IsConflict(err) {
			return fmt.Errorf("resource conflict (version mismatch): %w", err)
		}
		return fmt.Errorf("failed to update resource: %w", err)
	}
	return nil
}

// Patch patches a cluster-scoped resource
func (r *K8sClusterRepository[T, L]) Patch(ctx context.Context, name string, patchData map[string]interface{}) (T, error) {
	obj := r.newObject()
	key := types.NamespacedName{Name: name}

	// Get the existing resource
	if err := r.client.Get(ctx, key, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return obj, fmt.Errorf("resource not found %s: %w", name, err)
		}
		return obj, fmt.Errorf("failed to get resource for patch: %w", err)
	}

	// Convert patch data to JSON
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return obj, fmt.Errorf("failed to marshal patch data: %w", err)
	}

	// Apply merge patch
	patch := client.RawPatch(types.MergePatchType, patchBytes)
	if err := r.client.Patch(ctx, obj, patch); err != nil {
		if apierrors.IsNotFound(err) {
			return obj, fmt.Errorf("resource not found %s: %w", name, err)
		}
		return obj, fmt.Errorf("failed to patch resource: %w", err)
	}

	return obj, nil
}

// Delete deletes a cluster-scoped resource by name
func (r *K8sClusterRepository[T, L]) Delete(ctx context.Context, name string) error {
	obj := r.newObject()
	obj.SetName(name)

	if err := r.client.Delete(ctx, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("resource not found %s: %w", name, err)
		}
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

// List lists cluster-scoped resources
func (r *K8sClusterRepository[T, L]) List(ctx context.Context, opts ...ListOption) (L, error) {
	list := r.newList()

	// Apply options
	options := &ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Build client list options (no namespace for cluster-scoped)
	listOpts := &client.ListOptions{}

	if options.LabelSelector != "" {
		if selector, err := metav1.ParseToLabelSelector(options.LabelSelector); err == nil {
			listOpts.LabelSelector, _ = metav1.LabelSelectorAsSelector(selector)
		}
	}

	if options.Limit > 0 {
		listOpts.Limit = options.Limit
	}

	if options.Continue != "" {
		listOpts.Continue = options.Continue
	}

	if err := r.client.List(ctx, list, listOpts); err != nil {
		return list, fmt.Errorf("failed to list resources: %w", err)
	}

	return list, nil
}

// Watch watches for changes to cluster-scoped resources
func (r *K8sClusterRepository[T, L]) Watch(ctx context.Context, opts ...WatchOption) (<-chan WatchEvent[T], error) {
	// TODO: Implement watch functionality using controller-runtime client
	return nil, fmt.Errorf("watch functionality not yet implemented")
}

// K8sNamespacedRepository implements NamespacedRepository for namespace-scoped resources
type K8sNamespacedRepository[T client.Object, L client.ObjectList] struct {
	client    client.Client
	newObject func() T
	newList   func() L
}

// NewK8sNamespacedRepository creates a new namespace-scoped repository
func NewK8sNamespacedRepository[T client.Object, L client.ObjectList](
	k8sClient client.Client,
	newObject func() T,
	newList func() L,
) NamespacedRepository[T, L] {
	return &K8sNamespacedRepository[T, L]{
		client:    k8sClient,
		newObject: newObject,
		newList:   newList,
	}
}

// Create creates a new namespace-scoped resource
func (r *K8sNamespacedRepository[T, L]) Create(ctx context.Context, obj T) error {
	if err := r.client.Create(ctx, obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("resource already exists: %w", err)
		}
		return fmt.Errorf("failed to create resource: %w", err)
	}
	return nil
}

// Get retrieves a namespace-scoped resource by name and namespace
func (r *K8sNamespacedRepository[T, L]) Get(ctx context.Context, namespace, name string) (T, error) {
	obj := r.newObject()
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	if err := r.client.Get(ctx, key, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return obj, fmt.Errorf("resource not found %s/%s: %w", namespace, name, err)
		}
		return obj, fmt.Errorf("failed to get resource: %w", err)
	}

	return obj, nil
}

// Update updates a namespace-scoped resource
func (r *K8sNamespacedRepository[T, L]) Update(ctx context.Context, obj T) error {
	if err := r.client.Update(ctx, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("resource not found: %w", err)
		}
		if apierrors.IsConflict(err) {
			return fmt.Errorf("resource conflict (version mismatch): %w", err)
		}
		return fmt.Errorf("failed to update resource: %w", err)
	}
	return nil
}

// Patch patches a namespace-scoped resource
func (r *K8sNamespacedRepository[T, L]) Patch(ctx context.Context, namespace, name string, patchData map[string]interface{}) (T, error) {
	obj := r.newObject()
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	// Get the existing resource
	if err := r.client.Get(ctx, key, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return obj, fmt.Errorf("resource not found %s/%s: %w", namespace, name, err)
		}
		return obj, fmt.Errorf("failed to get resource for patch: %w", err)
	}

	// Convert patch data to JSON
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return obj, fmt.Errorf("failed to marshal patch data: %w", err)
	}

	// Apply merge patch
	patch := client.RawPatch(types.MergePatchType, patchBytes)
	if err := r.client.Patch(ctx, obj, patch); err != nil {
		if apierrors.IsNotFound(err) {
			return obj, fmt.Errorf("resource not found %s/%s: %w", namespace, name, err)
		}
		return obj, fmt.Errorf("failed to patch resource: %w", err)
	}

	return obj, nil
}

// Delete deletes a namespace-scoped resource by name and namespace
func (r *K8sNamespacedRepository[T, L]) Delete(ctx context.Context, namespace, name string) error {
	obj := r.newObject()
	obj.SetName(name)
	obj.SetNamespace(namespace)

	if err := r.client.Delete(ctx, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("resource not found %s/%s: %w", namespace, name, err)
		}
		return fmt.Errorf("failed to delete resource: %w", err)
	}
	return nil
}

// List lists namespace-scoped resources
func (r *K8sNamespacedRepository[T, L]) List(ctx context.Context, namespace string, opts ...ListOption) (L, error) {
	list := r.newList()

	// Apply options
	options := &ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Build client list options
	listOpts := &client.ListOptions{
		Namespace: namespace,
	}

	if options.LabelSelector != "" {
		if selector, err := metav1.ParseToLabelSelector(options.LabelSelector); err == nil {
			listOpts.LabelSelector, _ = metav1.LabelSelectorAsSelector(selector)
		}
	}

	if options.Limit > 0 {
		listOpts.Limit = options.Limit
	}

	if options.Continue != "" {
		listOpts.Continue = options.Continue
	}

	if err := r.client.List(ctx, list, listOpts); err != nil {
		return list, fmt.Errorf("failed to list resources: %w", err)
	}

	return list, nil
}

// Watch watches for changes to namespace-scoped resources
func (r *K8sNamespacedRepository[T, L]) Watch(ctx context.Context, namespace string, opts ...WatchOption) (<-chan WatchEvent[T], error) {
	// TODO: Implement watch functionality using controller-runtime client
	return nil, fmt.Errorf("watch functionality not yet implemented")
}