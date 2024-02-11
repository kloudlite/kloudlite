package functions

func MapMerge[K comparable, V any](sources ...map[K]V) map[K]V {
	m := make(map[K]V, 2*len(sources))
	for i := range sources {
		for k, v := range sources[i] {
			m[k] = v
		}
	}

	return m
}
