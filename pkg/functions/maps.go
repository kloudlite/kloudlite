package functions

// MapSet sets a key, value in a map. If a map is nil, it firsts initializes the map
func MapSet[T any](m *map[string]T, key string, value T) {
	if *m == nil {
		*m = make(map[string]T)
	}
	(*m)[key] = value
}

// MapContains checks if `destination` contains all keys from `source`
func MapContains[T comparable](destination map[string]T, source map[string]T) bool {
	if len(destination) == 0 && len(source) == 0 {
		return true
	}

	for k, v := range source {
		if destination[k] != v {
			return false
		}
	}
	return true
}

func MapEqual[K comparable, V comparable](first map[K]V, second map[K]V) bool {
	if len(first) != len(second) {
		return false
	}

	for k := range first {
		if second[k] != first[k] {
			return false
		}
	}
	return true
}

func MapHasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}
