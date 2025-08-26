package functions

import "slices"

func Contains[T comparable](arr []T, item T) bool {
	return slices.Contains(arr, item)
}

func ContainsAll[T comparable](arr []T, items []T) bool {
	m := make(map[T]bool, len(arr))
	for i := range arr {
		m[arr[i]] = true
	}
	for i := range items {
		if !m[items[i]] {
			return false
		}
	}
	return true
}

func ContainsAllWithPredicate[T comparable, K comparable](arr []T, items []T, predicate func(item T) K) bool {
	m := make(map[K]bool, len(arr))
	for i := range arr {
		m[predicate(arr[i])] = true
	}

	for i := range items {
		if !m[predicate(items[i])] {
			return false
		}
	}
	return true
}

func TransformSlice[T any, V any](items []T, predicate func(item T) V) []V {
	result := make([]V, 0, len(items))
	for i := range items {
		result = append(result, predicate(items[i]))
	}

	return result
}
