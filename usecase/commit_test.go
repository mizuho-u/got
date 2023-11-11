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

	if err := usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), f); err != nil {
		t.Fatal(err)
	}

	// act
	out := &bytes.Buffer{}
	err := usecase.Commit(newContext(dir, "Mizuho Ueda", "mi_ueda@u-m.dev", out, out), "First Commit.\n\nthe third and subsequent lines...\n", time.Unix(1694356071, 0))

	// assert
	if err != nil {
		t.Fatal(err)
	}

	expect := `[(root-commit) 489512179ae8ab55607b0e109221d2a38edacfca] First Commit.`

	if out.String() != expect {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out.String())
	}

}

func TestSecondCommit(t *testing.T) {

	// arrange
	dir := initDir(t)
	f := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	if err := usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), f); err != nil {
		t.Fatal(err)
	}

	err := usecase.Commit(newContext(dir, "Mizuho Ueda", "mi_ueda@u-m.dev", &bytes.Buffer{}, &bytes.Buffer{}), "First Commit.\n\nthe third and subsequent lines...\n", time.Unix(1697289936, 0))
	if err != nil {
		t.Fatal(err)
	}

	// act
	f2 := createFile(t, dir, "hello2.txt", []byte("Hello world 2.\n"))

	if err := usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), f2); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	err = usecase.Commit(newContext(dir, "Mizuho Ueda", "mi_ueda@u-m.dev", out, out), "second commit\n", time.Unix(1697289992, 0))
	if err != nil {
		t.Fatal(err)
	}

	// assert
	expect := `[3e69b36ae663a7361d6bdbdc154952aabdfe86f2] second commit`

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

	if err := usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), hello); err != nil {
		t.Fatal(err)
	}

	// act
	out := &bytes.Buffer{}
	err := usecase.Commit(newContext(dir, "Mizuho Ueda", "mi_ueda@u-m.dev", out, out), "commit a executable file\n", time.Unix(1697288601, 0))
	if err != nil {
		t.Fatal(err)
	}

	expect := `[(root-commit) 09edd72799f0ed3fc1350b19ce6eb3b64fabdc01] commit a executable file`

	if out.String() != expect {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out.String())
	}

}
