package internal

import "testing"

func TestQueur(t *testing.T) {

	q := Queue[string]{"aaa"}

	if got, err := q.Dequeue(); got != "aaa" || err != nil {
		t.Errorf("dequeue failed got %s expec %s", got, "aaa")
	}

	q.Enqueue("bbb")
	if got, err := q.Dequeue(); got != "bbb" || err != nil {
		t.Errorf("dequeue failed got %s expect %s", got, "bbb")
	}

}
