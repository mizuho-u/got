package internal

func MapE[A ~[]V, V any](m A, f func(V) (V, error)) ([]V, error) {

	r := make([]V, 0, len(m))
	for _, v := range m {

		v, err := f(v)
		if err != nil {
			return nil, err
		}

		r = append(r, v)
	}

	return r, nil
}

func Filter[A ~[]V, V any](a A, f func(V) bool) []V {

	r := make([]V, 0, len(a))
	for _, v := range a {
		if f(v) {
			r = append(r, v)
		}
	}

	return r
}
