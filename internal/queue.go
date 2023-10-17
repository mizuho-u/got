package internal

import "errors"

type Queue[T any] []T

func (q *Queue[T]) Enqueue(v T) {
	*q = append(*q, v)
}

func (q *Queue[T]) Dequeue() (ret T, err error) {

	if len(*q) == 0 {
		err = errors.New("failed to dequeue")
		return
	}

	ret = (*q)[0]
	*q = (*q)[1:]

	return
}
