package usecase_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func TestStatus(t *testing.T) {

	dir := initDir(t)
	f1 := createFile(t, dir, "hello.txt", []byte("hello.\n"))
	f2 := createFile(t, dir, "world.txt", []byte("world.\n"))

	out := &bytes.Buffer{}
	if code := usecase.Status(newContext(dir, "", "", out, out)); code != 0 {
		t.Fatal("expect exit code 0, got ", code)
	}

	expect := fmt.Sprintf("?? %s\n?? %s\n", filepath.Base(f1), filepath.Base(f2))

	if out.String() != expect {
		t.Fatalf("expect %s, got %s", expect, out)
	}

}

func TestStatusIndex(t *testing.T) {

	dir := initDir(t)
	f1 := createFile(t, dir, "hello.txt", []byte("hello.\n"))

	out := &bytes.Buffer{}
	if code := usecase.Add(newContext(dir, "", "", out, out), f1); code != 0 {
		t.Fatal(out)
	}

	out.Reset()
	if code := usecase.Commit(newContext(dir, "", "", out, out), "commit message", time.Unix(1677142145, 0)); code != 0 {
		t.Fatal(out)
	}

	f2 := createFile(t, dir, "world.txt", []byte("world.\n"))

	out.Reset()
	if code := usecase.Status(newContext(dir, "", "", out, out)); code != 0 {
		t.Fatal("expect exit code 0, got ", code)
	}

	expect := fmt.Sprintf("?? %s\n", filepath.Base(f2))
	if out.String() != expect {
		t.Fatalf("expect \n%s, got \n%s", expect, out)
	}

}

func TestStatusUntrackedDirectories(t *testing.T) {

	testt := []struct {
		description string
		files       []string
		dirs        []string
		tracked     []string
		expect      string
	}{
		{
			description: "lists untracked directories, not their contents",
			files:       []string{"file.txt", "dir/another.txt"},
			tracked:     []string{},
			expect:      "?? dir/\n?? file.txt\n",
		},
		{
			description: "lists untracked files inside tracked directories",
			files:       []string{"a/b/inner.txt", "a/outer.txt", "a/b/c/file.txt"},
			tracked:     []string{"a/b/inner.txt"},
			expect:      "?? a/b/c/\n?? a/outer.txt\n",
		},
		{
			description: "does not list empty untracked directories",
			dirs:        []string{"outer"},
			expect:      "",
		},
		// todo
		// {
		// 	description: "lists untracked directories that indirectly contain files",
		// 	files:       []string{"outer/inner/file.txt"},
		// 	expect:      "?? outer/\n",
		// },
		// {
		// 	description: "lists untracked directories that indirectly contain files",
		// 	files:       []string{"outer/file.txt", "outer/inner/file.txt"},
		// 	expect:      "?? outer/\n",
		// },
	}

	for _, tc := range testt {

		t.Run(tc.description, func(t *testing.T) {

			dir := initDir(t)

			for _, d := range tc.dirs {
				createDir(t, dir, d)
			}

			fnames := map[string]string{}
			for _, f := range tc.files {
				abs := createFile(t, dir, f, []byte("hello.\n"))
				fnames[f] = abs
			}

			for _, f := range tc.tracked {

				add(t, dir, fnames[f])
				commit(t, dir, "commit message", time.Unix(1677142145, 0))

			}

			out := &bytes.Buffer{}
			if code := usecase.Status(newContext(dir, "", "", out, out)); code != 0 {
				t.Error("expect exit code 0, got ", code)
			}

			if out.String() != tc.expect {
				t.Errorf("expect \n%s, got \n%s", tc.expect, out)
			}

		})

	}

}
