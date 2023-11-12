package repository

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mizuho-u/got/repository/internal"
)

func TestAddEntry(t *testing.T) {

	tt := []struct {
		description string
		add         []string
		expect      []string
	}{
		{
			description: "add a single entry",
			add:         []string{"hello.txt"},
			expect:      []string{"hello.txt"},
		},
		{
			description: "add a file whose parent directory had the same name as an existing file",
			add:         []string{"alice.txt", "bob.txt", "alice.txt/nested.txt"},
			expect:      []string{"alice.txt/nested.txt", "bob.txt"},
		},
		{
			description: "replace a directory with a file",
			add:         []string{"alice.txt", "nested/bob.txt", "nested"},
			expect:      []string{"alice.txt", "nested"},
		},
		{
			description: "recursively replcaes a directory with a file",
			add:         []string{"alice.txt", "nested/bob.txt", "nested/inner/claire.txt", "nested"},
			expect:      []string{"alice.txt", "nested"},
		},
	}

	for _, tc := range tt {

		t.Run(tc.description, func(t *testing.T) {

			index, _ := newIndex()

			for _, f := range tc.add {
				index.add(NewIndexEntry(f, "", &FileStat{}))
			}

			got := internal.Map(index.entries, func(v *indexEntry) string { return v.filename })

			if diff := cmp.Diff(got, tc.expect, cmpopts.SortSlices(func(i, j string) bool { return i < j })); diff != "" {
				t.Errorf("serialized index not match. %s", diff)
				return
			}

		})

	}

}
