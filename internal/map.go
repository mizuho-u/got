package internal

func Keys[M ~map[K]V, K comparable, V any](m M) []K {

	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}

	return r
}

func Map[A ~[]V, V any](m A, f func(V) V) []V {

	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, f(v))
	}

	return r
}
