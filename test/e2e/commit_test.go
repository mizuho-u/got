package e2e

import (
	"os"
	"os/exec"
	"regexp"
	"testing"
)

func TestFirstCommit(t *testing.T) {

	// arrange
	build := buildpath(t)
	tempdir := initDir(t, build)

	createFile(t, tempdir, "hello.txt", []byte("Hello world.\n"))

	// act
	cmd := `echo "First Commit.\n\nthe third and subsequent lines..." | ` + build + " -C " + tempdir + " commit"
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		t.Fatal("first commit failed ", string(out), err)
	}

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
	createFile(t, tempdir, "hello.txt", []byte("Hello world.\n"))

	cmd := `echo "First Commit.\n\nthe third and subsequent lines..." | ` + build + " -C " + tempdir + " commit"
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		t.Fatal("first commit failed. ", string(out), err)
	}

	// act
	cmd = `echo "second commit" | ` + build + " -C " + tempdir + " commit"
	out, err = exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		t.Fatal("second commit failed ", string(out))
	}

	// assert
	expect := `\[[0-9a-f]{40}\] second commit`

	if !regexp.MustCompile(expect).MatchString(string(out)) {
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

	// act
	cmd := `echo "commit a executable file" | ` + build + " -C " + tempdir + " commit"
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		t.Fatal("commit a file failed ", string(out), err)
	}

	// assert
	expect := `\[\(root-commit\) [0-9a-f]{40}\] commit a executable file`

	if !regexp.MustCompile(expect).MatchString(string(out)) {
		t.Fatalf("unexpected output. expect %s, got %s", expect, out)
	}

}
