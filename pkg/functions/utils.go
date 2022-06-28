package functions

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
