package model

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mizuho-u/got/model/internal"
)

func TestAddEntry(t *testing.T) {

	tt := []struct {
		description string
		files       []string
		expect      []string
	}{
		{
			description: "add a single entry",
			files:       []string{"hello.txt"},
			expect:      []string{"hello.txt"},
		},
		{
			description: "add a file whose parent directory had the same name as an existing file",
			files:       []string{"alice.txt", "bob.txt", "alice.txt/nested.txt"},
			expect:      []string{"alice.txt/nested.txt", "bob.txt"},
		},
	}

	for _, tc := range tt {

		t.Run(tc.description, func(t *testing.T) {

			index, _ := newIndex()

			for _, f := range tc.files {
				index.add(NewIndexEntry(f, "", &FileStat{}))
			}

			got := internal.Map(index.entries, func(v *indexEntry) string { return v.filename })

			if diff := cmp.Diff(got, tc.expect); diff != "" {
				t.Errorf("serialized index not match. %s", diff)
				return
			}

		})

	}

}
