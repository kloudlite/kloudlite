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
