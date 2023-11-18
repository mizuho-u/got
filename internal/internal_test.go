package internal_test

import (
	"testing"

	"github.com/mizuho-u/got/internal"
)

func TestSet(t *testing.T) {

	s := internal.NewSet[string]()
	s.Set("aaa")
	s.Set("bbb")
	s.Set("ccc")

	if s.Length() != 3 {
		t.Errorf("expect length 3, got %d", s.Length())
	}

	aaa, bbb, ccc := false, false, false
	for _, v := range s.Iter() {

		if v == "aaa" {
			aaa = true
		}

		if v == "bbb" {
			bbb = true
		}

		if v == "ccc" {
			ccc = true
		}

	}

	if !(aaa && bbb && ccc) {
		t.Errorf("aaa %t bbb %t ccc %t", aaa, bbb, ccc)
	}

}
