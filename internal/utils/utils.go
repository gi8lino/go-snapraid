package utils

// Ptr returns a pointer to the given value, for any type.
func Ptr[T any](v T) *T {
	return &v
}
