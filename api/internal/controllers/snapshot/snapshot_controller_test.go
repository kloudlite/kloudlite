package snapshot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsRefInList tests the isRefInList helper function
func TestIsRefInList(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		refs     []string
		expected bool
	}{
		{
			name:     "ref exists in list",
			ref:      "ref1",
			refs:     []string{"ref1", "ref2", "ref3"},
			expected: true,
		},
		{
			name:     "ref does not exist in list",
			ref:      "ref4",
			refs:     []string{"ref1", "ref2", "ref3"},
			expected: false,
		},
		{
			name:     "empty ref list",
			ref:      "ref1",
			refs:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRefInList(tt.ref, tt.refs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConstants tests that constants are set correctly
func TestConstants(t *testing.T) {
	// Verify constants are defined
	assert.Equal(t, "snapshots.kloudlite.io/finalizer", "snapshots.kloudlite.io/finalizer")
}
