package usecase_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mizuho-u/got/types"
	"github.com/mizuho-u/got/usecase"
)

func TestCheckout(t *testing.T) {

	testt := []struct {
		description string
		a           map[string][]byte
		b           map[string][]byte
	}{
		{
			description: "same file",
			a: map[string][]byte{
				"a.txt": []byte("a"),
			},
			b: map[string][]byte{
				"a.txt": []byte("a"),
			},
		},
		{
			description: "file contents changed",
			a: map[string][]byte{
				"a.txt": []byte("a"),
			},
			b: map[string][]byte{
				"a.txt": []byte("b"),
			},
		},
		{
			description: "change file to directory",
			a: map[string][]byte{
				"a/b.txt": []byte("a"),
			},
			b: map[string][]byte{
				"a": []byte("b"),
			},
		},
		{
			description: "files in directories",
			a: map[string][]byte{
				"a.txt":   []byte("a"),
				"b/c.txt": []byte("c"),
				"b/d.txt": []byte("d"),
			},
			b: map[string][]byte{
				"a.txt":   []byte("a"),
				"b/c.txt": []byte("c2"),
				"b/e.txt": []byte("e"),
			},
		},
	}

	for _, tc := range testt {
		t.Run(tc.description, func(t *testing.T) {

			dir := initDir(t)

			for path, data := range tc.a {
				add(t, dir, createFile(t, dir, path, data))
			}

			commit(t, dir, "", "", "commit a", time.Unix(1694356071, 0))

			for path, data := range tc.b {

				if exists(dir, path) {
					removeAll(t, dir, path)
				}

				add(t, dir, createFile(t, dir, path, data))
			}

			commit(t, dir, "", "", "commit b", time.Unix(1694356071, 0))

			rev, err := types.NewRevision("HEAD^")
			if err != nil {
				t.Fatal(err)
			}

			if err := usecase.Checkout(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), rev); err != nil {
				t.Fatal(err)
			}

			for path := range tc.a {
				if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
					t.Error(err)
				}
			}

			for path := range tc.b {

				if _, ok := tc.a[path]; ok {
					continue
				}

				if stat, err := os.Stat(filepath.Join(dir, path)); err == nil && !stat.IsDir() {
					t.Errorf("file %s still exists", path)
				}
			}

			out := &bytes.Buffer{}
			if err := usecase.Status(newContext(dir, "", "", out, &bytes.Buffer{}), false); err != nil {
				t.Fatal(err)
			}

			if out.String() != "nothing to commit, working tree clean" {
				t.Fatalf("unexpected status message %s", out)
			}

		})
	}

}
