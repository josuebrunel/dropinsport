package util

import (
	"fmt"
	"strconv"
)

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

func F64(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func F64Fmt(v float64, f string) string {
	return fmt.Sprintf(f, v)
}
