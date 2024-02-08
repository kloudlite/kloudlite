package functions

func Contains[T comparable](arr []T, item T) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
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

func MapMerge[K comparable, V any](source ...map[K]V) map[K]V {
	result := make(map[K]V)

	for _, m := range source {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
