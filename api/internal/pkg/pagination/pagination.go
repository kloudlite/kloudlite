package pagination

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// DefaultPageSize is the default number of items to return per page
	DefaultPageSize = int64(100)
	// MaxPageSize is the maximum number of items allowed per page
	MaxPageSize = int64(1000)
)

// ListOptions holds pagination options
type ListOptions struct {
	Limit        int64
	Continue     string
	ListOptions  []client.ListOption
}

// NewListOptions creates a new ListOptions with default values
func NewListOptions() *ListOptions {
	return &ListOptions{
		Limit: DefaultPageSize,
	}
}

// ApplyLimit applies the limit to the list options
func (o *ListOptions) ApplyLimit() []client.ListOption {
	if o.Limit <= 0 {
		o.Limit = DefaultPageSize
	} else if o.Limit > MaxPageSize {
		o.Limit = MaxPageSize
	}

	opts := append([]client.ListOption{}, o.ListOptions...)
	if o.Continue != "" {
		opts = append(opts, client.Continue(o.Continue))
	}
	opts = append(opts, client.Limit(o.Limit))

	return opts
}

// ListAll retrieves all items from the API server using pagination
// This is useful when you need to process all items but want to avoid
// loading too many items into memory at once
func ListAll(ctx context.Context, c client.Reader, list client.ObjectList, opts ...client.ListOption) error {
	// Use a smaller page size for list operations to avoid overloading the API server
	pageSize := DefaultPageSize
	var continueToken string

	// Get the underlying list to append items
	itemsAccessor, err := meta.ListAccessor(list)
	if err != nil {
		return fmt.Errorf("failed to get list accessor: %w", err)
	}

	// Store the initial items if any
	var initialItems []client.Object
	if items, err := meta.ExtractList(list); err == nil {
		for i := range items {
			if obj, ok := items[i].(client.Object); ok {
				initialItems = append(initialItems, obj)
			}
		}
	}

	// Paginate through all pages
	for {
		// Create list options with pagination
		paginatedOpts := []client.ListOption{}
		paginatedOpts = append(paginatedOpts, opts...)
		if continueToken != "" {
			paginatedOpts = append(paginatedOpts, client.Continue(continueToken))
		}
		paginatedOpts = append(paginatedOpts, client.Limit(pageSize))

		// List with pagination
		if err := c.List(ctx, list, paginatedOpts...); err != nil {
			return fmt.Errorf("failed to list items: %w", err)
		}

		// Get the continue token for the next page
		continueToken = itemsAccessor.GetContinue()
		if continueToken == "" {
			// No more pages
			break
		}
	}

	return nil
}

// ListWithPagination lists items with pagination support
// Returns the continue token for fetching the next page
func ListWithPagination(ctx context.Context, c client.Reader, list client.ObjectList, opts *ListOptions) (string, error) {
	if opts == nil {
		opts = NewListOptions()
	}

	// Apply pagination options
	paginatedOpts := opts.ApplyLimit()

	// List with pagination
	if err := c.List(ctx, list, paginatedOpts...); err != nil {
		return "", fmt.Errorf("failed to list items: %w", err)
	}

	// Get the continue token
	itemsAccessor, err := meta.ListAccessor(list)
	if err != nil {
		return "", fmt.Errorf("failed to get list accessor: %w", err)
	}

	return itemsAccessor.GetContinue(), nil
}

// PaginationResult holds the result of a paginated list operation
type PaginationResult struct {
	Items      []client.Object
	Continue   string
	HasMore    bool
	TotalCount int
}

// ListPaginated retrieves a single page of items with pagination support
func ListPaginated(ctx context.Context, c client.Reader, list client.ObjectList, opts *ListOptions) (*PaginationResult, error) {
	if opts == nil {
		opts = NewListOptions()
	}

	// Apply pagination options
	paginatedOpts := opts.ApplyLimit()

	// List with pagination
	if err := c.List(ctx, list, paginatedOpts...); err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	// Extract items
	items, err := meta.ExtractList(list)
	if err != nil {
		return nil, fmt.Errorf("failed to extract items: %w", err)
	}

	// Convert to client.Object slice
	result := make([]client.Object, 0, len(items))
	for i := range items {
		if obj, ok := items[i].(client.Object); ok {
			result = append(result, obj)
		}
	}

	// Get the continue token
	itemsAccessor, err := meta.ListAccessor(list)
	if err != nil {
		return nil, fmt.Errorf("failed to get list accessor: %w", err)
	}

	continueToken := itemsAccessor.GetContinue()
	hasMore := continueToken != ""

	return &PaginationResult{
		Items:      result,
		Continue:   continueToken,
		HasMore:    hasMore,
		TotalCount: len(result),
	}, nil
}

// ForEachPage iterates through all items page by page
// The callback receives a slice of items for the current page
// Returns an error if the callback returns an error or listing fails
func ForEachPage(ctx context.Context, c client.Reader, list client.ObjectList, pageSize int64, callback func([]client.Object) error) error {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	var continueToken string
	totalProcessed := 0

	for {
		// Create list options with pagination
		opts := []client.ListOption{}
		if continueToken != "" {
			opts = append(opts, client.Continue(continueToken))
		}
		opts = append(opts, client.Limit(pageSize))

		// List with pagination
		if err := c.List(ctx, list, opts...); err != nil {
			return fmt.Errorf("failed to list items (processed %d): %w", totalProcessed, err)
		}

		// Extract items
		items, err := meta.ExtractList(list)
		if err != nil {
			return fmt.Errorf("failed to extract items: %w", err)
		}

		// Convert to client.Object slice
		objects := make([]client.Object, 0, len(items))
		for i := range items {
			if obj, ok := items[i].(client.Object); ok {
				objects = append(objects, obj)
			}
		}

		// Call the callback
		if err := callback(objects); err != nil {
			return fmt.Errorf("callback failed (processed %d): %w", totalProcessed, err)
		}

		totalProcessed += len(objects)

		// Get the continue token
		itemsAccessor, err := meta.ListAccessor(list)
		if err != nil {
			return fmt.Errorf("failed to get list accessor: %w", err)
		}

		continueToken = itemsAccessor.GetContinue()
		if continueToken == "" {
			// No more pages
			break
		}

		// Reset the list for the next iteration
		if err := meta.SetList(list, []runtime.Object{}); err != nil {
			return fmt.Errorf("failed to reset list: %w", err)
		}
	}

	return nil
}

// GetRemainingItemCount returns the estimated number of remaining items
// This is only available if the server supports it
func GetRemainingItemCount(list client.ObjectList) (int64, error) {
	itemsAccessor, err := meta.ListAccessor(list)
	if err != nil {
		return 0, fmt.Errorf("failed to get list accessor: %w", err)
	}

	rc := itemsAccessor.GetRemainingItemCount()
	if rc == nil {
		return 0, nil
	}

	return *rc, nil
}
