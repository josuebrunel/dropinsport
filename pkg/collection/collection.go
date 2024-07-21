package collection

func Exists[T any](list []T, fn func(t T) bool) bool {
	var r bool
	for _, i := range list {
		if fn(i) {
			return true
		}
	}
	return r
}

func Transform[T any, V any](list []T, fn func(T) V) []V {
	result := make([]V, len(list))
	for i, v := range list {
		result[i] = fn(v)
	}
	return result
}

func ZipApply[X any, Y any, R any](xx []X, yy []Y, fn func(X, Y) R) []R {
	result := make([]R, len(xx))
	for i, x := range xx {
		result[i] = fn(x, yy[i])
	}
	return result
}

func ToMap[T any, K comparable, V any](list []T, fn func(T) (K, V)) map[K]V {
	result := make(map[K]V)
	for _, i := range list {
		k, v := fn(i)
		result[k] = v
	}
	return result
}
