package util

func Deref[T any](t *T) T {
	var r T
	if t != nil {
		r = *t
	}
	return r
}
