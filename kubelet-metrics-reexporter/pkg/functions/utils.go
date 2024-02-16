package functions

func New[T any](x T) *T {
	return &x
}
