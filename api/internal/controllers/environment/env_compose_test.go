package environment

import (
	"testing"
)

// TestMakeStringSet tests the makeStringSet helper
func TestMakeStringSet(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected map[string]bool
	}{
		{
			name:     "empty slice",
			items:    []string{},
			expected: map[string]bool{},
		},
		{
			name:     "single item",
			items:    []string{"item1"},
			expected: map[string]bool{"item1": true},
		},
		{
			name:     "multiple items",
			items:    []string{"item1", "item2", "item3"},
			expected: map[string]bool{"item1": true, "item2": true, "item3": true},
		},
		{
			name:     "duplicate items",
			items:    []string{"item1", "item1", "item2"},
			expected: map[string]bool{"item1": true, "item2": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeStringSet(tt.items)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s: %v, got %v", k, v, result[k])
				}
			}
		})
	}
}
