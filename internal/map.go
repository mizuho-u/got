package internal

func Keys[M ~map[K]V, K comparable, V any](m M) []K {

	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}

	return r
}

func Map[M ~map[K]V, K comparable, V, R any](m M, f func(V) R) []R {

	r := make([]R, 0, len(m))
	for _, v := range m {
		r = append(r, f(v))
	}

	return r
}
