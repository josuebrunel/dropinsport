package util

import "fmt"

func Deref[T any](t *T) T {
	var r T
	if t != nil {
		r = *t
	}
	return r
}

func AssertType[T any](val any) T {
	var r T
	if v, ok := val.(T); ok {
		r = v
	}
	return r
}

func F64Fmt(v float64, f string) string {
	return fmt.Sprintf(f, v)
}
