package e2e

import (
	"os"
	"regexp"
	"testing"
)

func TestFirstCommit(t *testing.T) {

	// arrange
	build := buildpath(t)
	tempdir := initDir(t, build)

	f := createFile(t, tempdir, "hello.txt", []byte("Hello world.\n"))
	executeCmd(t, build+" -C "+tempdir+" add "+f)

	// act
	out := executeCmd(t, `echo "First Commit.\n\nthe third and subsequent lines..." | `+build+" -C "+tempdir+" commit")

	// assert
	expect := `\[\(root-commit\) [0-9a-f]{40}\] First Commit.`
	if !regexp.MustCompile(expect).MatchString(string(out)) {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out)
	}

}

func TestSecondCommit(t *testing.T) {

	// arrange
	build := buildpath(t)
	tempdir := initDir(t, build)

	f1 := createFile(t, tempdir, "hello.txt", []byte("Hello world.\n"))
	executeCmd(t, build+" -C "+tempdir+" add "+f1)

	executeCmd(t, `echo "First Commit.\n\nthe third and subsequent lines..." | `+build+" -C "+tempdir+" commit")

	// act
	f2 := createFile(t, tempdir, "hello2.txt", []byte("Hello world.\n"))
	executeCmd(t, build+" -C "+tempdir+" add "+f2)

	out := executeCmd(t, `echo "second commit" | `+build+" -C "+tempdir+" commit")

	// assert
	expect := `\[[0-9a-f]{40}\] second commit`
	if !regexp.MustCompile(expect).MatchString(out) {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out)
	}

}

func TestCommitExcutableFiles(t *testing.T) {

	// arrange
	build := buildpath(t)
	tempdir := initDir(t, build)

	hello := createFile(t, tempdir, "hello.txt", []byte("Hello world.\n"))
	if err := os.Chmod(hello, 0755); err != nil {
		t.Fatal("failed to chmod test file. ", err)
	}

	executeCmd(t, build+" -C "+tempdir+" add "+hello)

	// act
	out := executeCmd(t, `echo "commit a executable file" | `+build+" -C "+tempdir+" commit")

	// assert
	expect := `\[\(root-commit\) [0-9a-f]{40}\] commit a executable file`
	if !regexp.MustCompile(expect).MatchString(out) {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out)
	}

}
