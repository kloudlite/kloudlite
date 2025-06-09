package functions

func Reduce[T any, V any](items []T, reducerFn func(V, T), value V) V {
	for i := range items {
		item := items[i]
		reducerFn(value, item)
	}

	return value
}
