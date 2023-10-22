package usecase_test

import (
	"bytes"
	"fmt"
	"io/fs"
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
		{
			description: "lists untracked directories that indirectly contain files",
			files:       []string{"outer/file.txt", "outer/inner/file.txt", "outer2/inner2/innerinnner2/file2.txt"},
			expect:      "?? outer/\n?? outer2/\n",
		},
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

func TestStatusChangedContents(t *testing.T) {

	testt := []struct {
		description string
		newcontents map[string]string
		newmode     map[string]fs.FileMode
		newtimes    map[string]int64
		delete      []string
		expect      string
	}{
		{
			description: "prints nothing when no files are changed",
			newcontents: map[string]string{},
			expect:      "",
		},
		{
			description: "prints files with changed contents",
			newcontents: map[string]string{"1.txt": "changed", "a/2.txt": "modified"},
			expect:      " M 1.txt\n M a/2.txt\n",
		},
		{
			description: "reports files with changed modes",
			newmode:     map[string]fs.FileMode{"a/2.txt": 0755},
			expect:      " M a/2.txt\n",
		},
		{
			description: "reports modified files with unchanged size",
			newcontents: map[string]string{"a/b/3.txt": "hello"},
			expect:      " M a/b/3.txt\n",
		},
		{
			description: "reports modified files with unchanged size",
			newcontents: map[string]string{"a/b/3.txt": "hello"},
			expect:      " M a/b/3.txt\n",
		},
		{
			description: "prints nothing if a file is touched",
			newtimes:    map[string]int64{"a/b/3.txt": 1},
			expect:      "",
		},
		{
			description: "reports deleted files",
			delete:      []string{"a/2.txt"},
			expect:      " D a/2.txt\n",
		},
		{
			description: "reports files in deleted directories",
			delete:      []string{"a"},
			expect:      " D a/2.txt\n D a/b/3.txt\n",
		},
	}

	for _, tc := range testt {

		t.Run(tc.description, func(t *testing.T) {

			now := time.Now()

			dir := initDir(t)

			f1 := createFile(t, dir, "1.txt", []byte("one"))
			modifyFileTime(t, dir, "1.txt", now, now)

			f2 := createFile(t, dir, "a/2.txt", []byte("two"))
			modifyFileTime(t, dir, "a/2.txt", now, now)

			f3 := createFile(t, dir, "a/b/3.txt", []byte("three"))
			modifyFileTime(t, dir, "a/b/3.txt", now, now)

			add(t, dir, f1)
			add(t, dir, f2)
			add(t, dir, f3)

			commit(t, dir, "commit massage", time.Unix(1677142145, 0))

			for file, contents := range tc.newcontents {
				createFile(t, dir, file, []byte(contents))
			}

			for file, mode := range tc.newmode {
				modifyFileMode(t, dir, file, mode)
			}

			for _, name := range tc.delete {
				removeAll(t, dir, name)
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

func TestStatusHeadIndexDifferences(t *testing.T) {

	testt := []struct {
		description      string
		newAddedContents map[string]string
		expect           string
	}{
		{
			description:      "reports a file added to a tracked directory",
			newAddedContents: map[string]string{"a/4.txt": "four"},
			expect:           "A  a/4.txt\n",
		},
		{
			description:      "prints files with changed contents",
			newAddedContents: map[string]string{"d/e/5.txt": "five"},
			expect:           "A  d/e/5.txt\n",
		},
	}

	for _, tc := range testt {

		t.Run(tc.description, func(t *testing.T) {

			now := time.Now()

			dir := initDir(t)

			f1 := createFile(t, dir, "1.txt", []byte("one"))
			modifyFileTime(t, dir, "1.txt", now, now)
			add(t, dir, f1)

			f2 := createFile(t, dir, "a/2.txt", []byte("two"))
			modifyFileTime(t, dir, "a/2.txt", now, now)
			add(t, dir, f2)

			f3 := createFile(t, dir, "a/b/3.txt", []byte("three"))
			modifyFileTime(t, dir, "a/b/3.txt", now, now)
			add(t, dir, f3)

			commit(t, dir, "commit massage", time.Unix(1677142145, 0))

			for file, contents := range tc.newAddedContents {
				f := createFile(t, dir, file, []byte(contents))
				add(t, dir, f)
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
