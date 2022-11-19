package functions

import (
	"testing"
)

type mapContainsArgs[T comparable] struct {
	target map[string]T
	m      map[string]T
}

func TestMapContains(t *testing.T) {
	tests := []struct {
		name string
		args mapContainsArgs[string]
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "dest is nil and source is nil",
			args: mapContainsArgs[string]{nil, nil},
			want: true,
		},
		{
			name: "dest is nil and source is empty",
			args: mapContainsArgs[string]{nil, map[string]string{}},
			want: true,
		},
		{
			name: "dest is empty and source is nil",
			args: mapContainsArgs[string]{map[string]string{}, nil},
			want: true,
		},
		{
			name: "dest is nil and source is not empty",
			args: mapContainsArgs[string]{nil, map[string]string{"hello": "world"}},
			want: false,
		},
		{
			name: "dest is empty and source is not empty",
			args: mapContainsArgs[string]{map[string]string{}, map[string]string{"hello": "world"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := MapContains(tt.args.target, tt.args.m); got != tt.want {
					t.Errorf("MapContains() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
