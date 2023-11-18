package internal

type Set[T comparable] interface {
	Set(v T)
	Merge(other Set[T])
	Length() int
	Iter() []T
	Has(v T) bool
}

type set[T comparable] map[T]interface{}

func NewSet[T comparable]() Set[T] {
	return set[T]{}
}

func NewSetFromArray[T comparable](a []T) Set[T] {

	s := set[T]{}

	for _, v := range a {
		s.Set(v)
	}

	return s
}

func (s set[T]) Set(v T) {
	s[v] = nil
}

func (s set[T]) Has(v T) bool {
	_, ok := s[v]

	return ok
}

func (s set[T]) Merge(other Set[T]) {

	for _, v := range other.Iter() {
		s.Set(v)
	}
}

func (s set[T]) Length() int {
	return len(s)
}

func (s set[T]) Iter() []T {
	return Keys(s)
}
