package repository

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Repository defines the common interface for all resource repositories
type Repository[T client.Object, L client.ObjectList] interface {
	// Create creates a new resource
	Create(ctx context.Context, obj T) error

	// Get retrieves a resource by name and namespace
	Get(ctx context.Context, name, namespace string) (T, error)

	// Update updates an existing resource
	Update(ctx context.Context, obj T) error

	// Delete deletes a resource by name and namespace
	Delete(ctx context.Context, name, namespace string) error

	// List lists resources with optional filters
	List(ctx context.Context, namespace string, opts ...ListOption) (L, error)

	// Watch watches for changes to resources (optional - may not be implemented by all repositories)
	Watch(ctx context.Context, namespace string, opts ...WatchOption) (<-chan WatchEvent[T], error)
}

// ListOption defines options for list operations
type ListOption func(*ListOptions)

// ListOptions contains options for listing resources
type ListOptions struct {
	LabelSelector string
	FieldSelector string
	Limit         int64
	Continue      string
}

// WithLabelSelector adds a label selector to list options
func WithLabelSelector(selector string) ListOption {
	return func(opts *ListOptions) {
		opts.LabelSelector = selector
	}
}

// WithFieldSelector adds a field selector to list options
func WithFieldSelector(selector string) ListOption {
	return func(opts *ListOptions) {
		opts.FieldSelector = selector
	}
}

// WithLimit adds a limit to list options
func WithLimit(limit int64) ListOption {
	return func(opts *ListOptions) {
		opts.Limit = limit
	}
}

// WithContinue adds a continue token to list options
func WithContinue(continueToken string) ListOption {
	return func(opts *ListOptions) {
		opts.Continue = continueToken
	}
}

// WatchOption defines options for watch operations
type WatchOption func(*WatchOptions)

// WatchOptions contains options for watching resources
type WatchOptions struct {
	LabelSelector    string
	FieldSelector    string
	ResourceVersion  string
	TimeoutSeconds   *int64
	AllowWatchBookmarks bool
}

// WithWatchLabelSelector adds a label selector to watch options
func WithWatchLabelSelector(selector string) WatchOption {
	return func(opts *WatchOptions) {
		opts.LabelSelector = selector
	}
}

// WithWatchFieldSelector adds a field selector to watch options
func WithWatchFieldSelector(selector string) WatchOption {
	return func(opts *WatchOptions) {
		opts.FieldSelector = selector
	}
}

// WithResourceVersion sets the resource version for watch
func WithResourceVersion(version string) WatchOption {
	return func(opts *WatchOptions) {
		opts.ResourceVersion = version
	}
}

// WithWatchTimeout sets the timeout for watch operations
func WithWatchTimeout(seconds int64) WatchOption {
	return func(opts *WatchOptions) {
		opts.TimeoutSeconds = &seconds
	}
}

// WatchEventType represents the type of watch event
type WatchEventType string

const (
	WatchEventAdded    WatchEventType = "ADDED"
	WatchEventModified WatchEventType = "MODIFIED"
	WatchEventDeleted  WatchEventType = "DELETED"
	WatchEventError    WatchEventType = "ERROR"
)

// WatchEvent represents a watch event
type WatchEvent[T client.Object] struct {
	Type   WatchEventType
	Object T
	Error  error
}

// buildListOptions creates metav1.ListOptions from repository ListOptions
func buildListOptions(opts *ListOptions) metav1.ListOptions {
	listOpts := metav1.ListOptions{}

	if opts.LabelSelector != "" {
		listOpts.LabelSelector = opts.LabelSelector
	}
	if opts.FieldSelector != "" {
		listOpts.FieldSelector = opts.FieldSelector
	}
	if opts.Limit > 0 {
		listOpts.Limit = opts.Limit
	}
	if opts.Continue != "" {
		listOpts.Continue = opts.Continue
	}

	return listOpts
}

// buildWatchOptions creates metav1.ListOptions for watching from repository WatchOptions
func buildWatchOptions(opts *WatchOptions) metav1.ListOptions {
	watchOpts := metav1.ListOptions{
		Watch: true,
	}

	if opts.LabelSelector != "" {
		watchOpts.LabelSelector = opts.LabelSelector
	}
	if opts.FieldSelector != "" {
		watchOpts.FieldSelector = opts.FieldSelector
	}
	if opts.ResourceVersion != "" {
		watchOpts.ResourceVersion = opts.ResourceVersion
	}
	if opts.TimeoutSeconds != nil {
		watchOpts.TimeoutSeconds = opts.TimeoutSeconds
	}
	watchOpts.AllowWatchBookmarks = opts.AllowWatchBookmarks

	return watchOpts
}