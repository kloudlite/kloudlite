package functions

import (
	"strings"

	"github.com/gobuffalo/flect"
)

func ParseOnlyOption[T any](item []T) *T {
	if len(item) > 0 {
		return &item[0]
	}
	return nil
}

func NewTypeFromPointer[T any]() T {
	t := make([]T, 1)
	return t[0]
}

func First[T any](items []T) T {
	if len(items) > 0 {
		return items[0]
	}
	return *new(T)
}

func DefaultIfNil[T any](v *T, defaultVal ...T) T {
	if v == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return *new(T)
	}
	return *v
}

func New[T any](v T) *T {
	return &v
}

// IsNil is useful when checking for a typed pointer
// for primitive types, use v == nil
func IsNil[T any](v T) bool {
	var x T
	return func(a, b any) bool {
		return a == b
	}(v, x)
}

// RegularPlural is copied from https://github.com/kubernetes-sigs/kubebuilder/blob/afce6a0e8c2a6d5682be07bbe502e728dd619714/pkg/model/resource/utils.go#L71
func RegularPlural(singular string) string {
	return flect.Pluralize(strings.ToLower(singular))
}
