package functions

func Ptr[T any](v T) *T {
	return &v
}

func ValueOf[T any](v *T) T {
	if v == nil {
		var x T
		return x
	}
	return *v
}
