package internal

type Tuple[T1, T2 any] interface {
	Item1() T1
	Item2() T2
}

type tuple[T1, T2 any] struct {
	item1 T1
	item2 T2
}

func (t *tuple[T1, T2]) Item1() T1 {
	return t.item1
}

func (t *tuple[T1, T2]) Item2() T2 {
	return t.item2
}

func NewTuple[T1, T2 any](item1 T1, item2 T2) *tuple[T1, T2] {
	return &tuple[T1, T2]{item1, item2}
}
