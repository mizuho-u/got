package usecase_test

import (
	"testing"
	"time"
)

func TestFirstCommit(t *testing.T) {

	// arrange
	dir := initDir(t)
	f := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	add(t, dir, f)

	// act
	out := commit(t, dir, "Mizuho Ueda", "mi_ueda@u-m.dev", "First Commit.\n\nthe third and subsequent lines...\n", time.Unix(1694356071, 0))

	// assert
	expect := `[(root-commit) 489512179ae8ab55607b0e109221d2a38edacfca] First Commit.`
	if out != expect {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out)
	}

}

func TestSecondCommit(t *testing.T) {

	// arrange
	dir := initDir(t)

	f := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))
	add(t, dir, f)

	commit(t, dir, "Mizuho Ueda", "mi_ueda@u-m.dev", "First Commit.\n\nthe third and subsequent lines...\n", time.Unix(1697289936, 0))

	// act
	f2 := createFile(t, dir, "hello2.txt", []byte("Hello world 2.\n"))
	add(t, dir, f2)

	out := commit(t, dir, "Mizuho Ueda", "mi_ueda@u-m.dev", "second commit\n", time.Unix(1697289992, 0))

	// assert
	expect := `[3e69b36ae663a7361d6bdbdc154952aabdfe86f2] second commit`
	if out != expect {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out)
	}

}

func TestCommitExcutableFiles(t *testing.T) {

	// arrange
	dir := initDir(t)

	hello := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))
	modifyFileMode(t, "", hello, 0755)
	add(t, dir, hello)

	// act
	out := commit(t, dir, "Mizuho Ueda", "mi_ueda@u-m.dev", "commit a executable file\n", time.Unix(1697288601, 0))

	expect := `[(root-commit) 09edd72799f0ed3fc1350b19ce6eb3b64fabdc01] commit a executable file`
	if out != expect {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out)
	}

}
