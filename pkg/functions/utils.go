package functions

func ParseOnlyOption[T any](item []T) *T {
	if len(item) > 0 {
		return &item[0]
	}
	return nil
}

func New[T any]() T {
	t := make([]T, 1)
	return t[0]
}

func First[T any](items []T) T {
	if len(items) > 0 {
		return items[0]
	}
	return *new(T)
}
