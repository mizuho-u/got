package types_test

import (
	"testing"

	"github.com/mizuho-u/got/types"
)

func TestBranchName(t *testing.T) {

	testt := []struct {
		name  string
		valid bool
	}{
		{"topic", true},
		{"t=opic", true},
		{".topic", false},
		{"to/.pic", false},
		{"topic.lock", false},
		{"to..pic", false},
		{"top@{ic", false},
		{"to  pic", false},
	}

	for _, tc := range testt {

		_, err := types.NewBranchName(tc.name)
		if tc.valid && err != nil {
			t.Fatalf("valid branch name, but got error %s", err)
		}

		if !tc.valid && err == nil {
			t.Fatal("invalid branch name, but got no error")
		}

	}

}
