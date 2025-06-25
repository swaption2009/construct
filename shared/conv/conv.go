package conv

func Ptr[T any](v T) *T {
	return &v
}

func FromPtr[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}