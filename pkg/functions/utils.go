package functions

func ParseOnlyOption[T any](item []T) *T {
	if len(item) > 0 {
		return &item[0]
	}
	return nil
}
