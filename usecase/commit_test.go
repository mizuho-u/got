package usecase_test

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func TestFirstCommit(t *testing.T) {

	// arrange
	dir := initDir(t)
	f := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f); err != nil {
		t.Fatal("failed to add. ", err)
	}

	// act
	out := &bytes.Buffer{}
	err := usecase.Commit(newContext(dir, out, out), "First Commit.\n\nthe third and subsequent lines...", time.Unix(1677142145, 0))

	// assert
	if err != nil {
		t.Fatal("failed to commit. ", err)
	}

	expect := `[(root-commit) 0be97431ca5456627193eda08dc0a7d0267045a5] First Commit.`

	if out.String() != expect {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out.String())
	}

}

func TestSecondCommit(t *testing.T) {

	// arrange
	dir := initDir(t)
	f := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f); err != nil {
		t.Fatal("failed to add. ", err)
	}

	err := usecase.Commit(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), "First Commit.\n\nthe third and subsequent lines...", time.Unix(1677142145, 0))
	if err != nil {
		t.Fatal("first commit failed. ", err)
	}

	// act
	f2 := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f2); err != nil {
		t.Fatal("failed to add. ", err)
	}

	out := &bytes.Buffer{}
	err = usecase.Commit(newContext(dir, out, out), "second commit", time.Unix(1677142145, 0))
	if err != nil {
		t.Fatal("second commit failed ", err)
	}

	// assert
	expect := `[e4c1b779b51993f90ac7808726920b4e7139f94c] second commit`

	if out.String() != expect {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out.String())
	}

}

func TestCommitExcutableFiles(t *testing.T) {

	// arrange
	dir := initDir(t)
	hello := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	if err := os.Chmod(hello, 0755); err != nil {
		t.Fatal("failed to chmod test file. ", err)
	}

	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), hello); err != nil {
		t.Fatal("failed to add. ", err)
	}

	// act
	out := &bytes.Buffer{}
	err := usecase.Commit(newContext(dir, out, out), "commit a executable file", time.Unix(1677142145, 0))

	// assert
	if err != nil {
		t.Fatal("failed to commit. ", err)
	}

	expect := `[(root-commit) fda0b416a0336b1b34339191a3827d22d2144c17] commit a executable file`

	if out.String() != expect {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out.String())
	}

}
