package repository

import (
	"context"
	"encoding/json"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sClusterRepository implements ClusterRepository for cluster-scoped resources
type K8sClusterRepository[T client.Object, L client.ObjectList] struct {
	client    client.WithWatch
	newObject func() T
	newList   func() L
}

// NewK8sClusterRepository creates a new cluster-scoped repository
func NewK8sClusterRepository[T client.Object, L client.ObjectList](
	k8sClient client.WithWatch,
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

	// Build client list options using helper function
	metav1ListOpts := buildListOptions(options)

	// Convert to client.ListOptions
	listOpts := &client.ListOptions{}
	if metav1ListOpts.LabelSelector != "" {
		if selector, err := metav1.ParseToLabelSelector(metav1ListOpts.LabelSelector); err == nil {
			listOpts.LabelSelector, _ = metav1.LabelSelectorAsSelector(selector)
		}
	}
	if metav1ListOpts.FieldSelector != "" {
		if fieldSelector, err := fields.ParseSelector(metav1ListOpts.FieldSelector); err == nil {
			listOpts.FieldSelector = fieldSelector
		}
	}
	if metav1ListOpts.Limit > 0 {
		listOpts.Limit = metav1ListOpts.Limit
	}
	if metav1ListOpts.Continue != "" {
		listOpts.Continue = metav1ListOpts.Continue
	}

	if err := r.client.List(ctx, list, listOpts); err != nil {
		return list, fmt.Errorf("failed to list resources: %w", err)
	}

	return list, nil
}

// Watch watches for changes to cluster-scoped resources
func (r *K8sClusterRepository[T, L]) Watch(ctx context.Context, opts ...WatchOption) (<-chan WatchEvent[T], error) {
	watchOpts := &WatchOptions{}
	for _, opt := range opts {
		opt(watchOpts)
	}

	// Build list options for watch
	list := r.newList()
	listOpts := &client.ListOptions{}

	if watchOpts.LabelSelector != "" {
		selector, err := metav1.ParseToLabelSelector(watchOpts.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to parse label selector: %w", err)
		}
		labelSelector, err := metav1.LabelSelectorAsSelector(selector)
		if err != nil {
			return nil, fmt.Errorf("failed to convert label selector: %w", err)
		}
		listOpts.LabelSelector = labelSelector
	}

	if watchOpts.FieldSelector != "" {
		listOpts.FieldSelector = fields.ParseSelectorOrDie(watchOpts.FieldSelector)
	}

	// Start the watch
	watcher, err := r.client.Watch(ctx, list, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to start watch: %w", err)
	}

	// Create event channel
	eventChan := make(chan WatchEvent[T], 100)

	// Start goroutine to process watch events
	go func() {
		defer close(eventChan)
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.ResultChan():
				if !ok {
					return
				}

				watchEvent := WatchEvent[T]{}

				switch event.Type {
				case watch.Added:
					watchEvent.Type = WatchEventAdded
				case watch.Modified:
					watchEvent.Type = WatchEventModified
				case watch.Deleted:
					watchEvent.Type = WatchEventDeleted
				case watch.Error:
					watchEvent.Type = WatchEventError
					if status, ok := event.Object.(*metav1.Status); ok {
						watchEvent.Error = fmt.Errorf("watch error: %s", status.Message)
					}
					eventChan <- watchEvent
					continue
				default:
					continue
				}

				if obj, ok := event.Object.(T); ok {
					watchEvent.Object = obj
					eventChan <- watchEvent
				}
			}
		}
	}()

	return eventChan, nil
}

// K8sNamespacedRepository implements NamespacedRepository for namespace-scoped resources
type K8sNamespacedRepository[T client.Object, L client.ObjectList] struct {
	client    client.WithWatch
	newObject func() T
	newList   func() L
}

// NewK8sNamespacedRepository creates a new namespace-scoped repository
func NewK8sNamespacedRepository[T client.Object, L client.ObjectList](
	k8sClient client.WithWatch,
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

	// Build client list options using helper function
	metav1ListOpts := buildListOptions(options)

	// Convert to client.ListOptions
	listOpts := &client.ListOptions{
		Namespace: namespace,
	}
	if metav1ListOpts.LabelSelector != "" {
		if selector, err := metav1.ParseToLabelSelector(metav1ListOpts.LabelSelector); err == nil {
			listOpts.LabelSelector, _ = metav1.LabelSelectorAsSelector(selector)
		}
	}
	if metav1ListOpts.FieldSelector != "" {
		if fieldSelector, err := fields.ParseSelector(metav1ListOpts.FieldSelector); err == nil {
			listOpts.FieldSelector = fieldSelector
		}
	}
	if metav1ListOpts.Limit > 0 {
		listOpts.Limit = metav1ListOpts.Limit
	}
	if metav1ListOpts.Continue != "" {
		listOpts.Continue = metav1ListOpts.Continue
	}

	if err := r.client.List(ctx, list, listOpts); err != nil {
		return list, fmt.Errorf("failed to list resources: %w", err)
	}

	return list, nil
}

// Watch watches for changes to namespace-scoped resources
func (r *K8sNamespacedRepository[T, L]) Watch(ctx context.Context, namespace string, opts ...WatchOption) (<-chan WatchEvent[T], error) {
	watchOpts := &WatchOptions{}
	for _, opt := range opts {
		opt(watchOpts)
	}

	// Build list options for watch
	list := r.newList()
	listOpts := &client.ListOptions{
		Namespace: namespace,
	}

	if watchOpts.LabelSelector != "" {
		selector, err := metav1.ParseToLabelSelector(watchOpts.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to parse label selector: %w", err)
		}
		labelSelector, err := metav1.LabelSelectorAsSelector(selector)
		if err != nil {
			return nil, fmt.Errorf("failed to convert label selector: %w", err)
		}
		listOpts.LabelSelector = labelSelector
	}

	if watchOpts.FieldSelector != "" {
		listOpts.FieldSelector = fields.ParseSelectorOrDie(watchOpts.FieldSelector)
	}

	// Start the watch
	watcher, err := r.client.Watch(ctx, list, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to start watch: %w", err)
	}

	// Create event channel
	eventChan := make(chan WatchEvent[T], 100)

	// Start goroutine to process watch events
	go func() {
		defer close(eventChan)
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.ResultChan():
				if !ok {
					return
				}

				watchEvent := WatchEvent[T]{}

				switch event.Type {
				case watch.Added:
					watchEvent.Type = WatchEventAdded
				case watch.Modified:
					watchEvent.Type = WatchEventModified
				case watch.Deleted:
					watchEvent.Type = WatchEventDeleted
				case watch.Error:
					watchEvent.Type = WatchEventError
					if status, ok := event.Object.(*metav1.Status); ok {
						watchEvent.Error = fmt.Errorf("watch error: %s", status.Message)
					}
					eventChan <- watchEvent
					continue
				default:
					continue
				}

				if obj, ok := event.Object.(T); ok {
					watchEvent.Object = obj
					eventChan <- watchEvent
				}
			}
		}
	}()

	return eventChan, nil
}
