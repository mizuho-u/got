package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestAdd(t *testing.T) {

	build := buildpath(t)
	tempdir := t.TempDir()

	_, err := exec.Command(build, "init", tempdir).Output()
	if err != nil {
		t.Fatal("init repository failed ", err)
	}

	f := createFile(t, tempdir, "hello.txt", []byte("Hello world.\n"))

	out, err := exec.Command(build, "-C", tempdir, "add", f).CombinedOutput()
	if err != nil {
		t.Fatal("add a file failed ", string(out))
	}

	testlsfiles(t, tempdir, "hello.txt\n")

}

func TestAddMultipleFiles(t *testing.T) {

	// arrange
	build := buildpath(t)
	tempdir := t.TempDir()

	_, err := exec.Command(build, "init", tempdir).Output()
	if err != nil {
		t.Fatal("init repository failed ", err)
	}

	f1 := createFile(t, tempdir, "hello.txt", []byte("hello.\n"))
	f2 := createFile(t, tempdir, "world.txt", []byte("world.\n"))

	// act
	out, err := exec.Command(build, "-C", tempdir, "add", f1, f2).CombinedOutput()
	if err != nil {
		t.Fatal("add a file failed ", string(out))
	}

	// assert
	testlsfiles(t, tempdir, "hello.txt\nworld.txt\n")

}

func TestAddFilesFromDirectory(t *testing.T) {

	// arrange
	build := buildpath(t)
	tempdir := t.TempDir()

	_, err := exec.Command(build, "init", tempdir).Output()
	if err != nil {
		t.Fatal("init repository failed ", err)
	}

	createFile(t, tempdir, "hello.txt", []byte("hello.\n"))
	createFile(t, tempdir, "world.txt", []byte("world.\n"))

	// act
	out, err := exec.Command(build, "-C", tempdir, "add", tempdir).CombinedOutput()
	if err != nil {
		t.Fatal("add a file failed ", string(out))
	}

	// assert
	testlsfiles(t, tempdir, "hello.txt\nworld.txt\n")

}

func TestModifyTheIndex(t *testing.T) {

	// arrange
	build := buildpath(t)
	tempdir := t.TempDir()

	_, err := exec.Command(build, "init", tempdir).Output()
	if err != nil {
		t.Fatal("init repository failed ", err)
	}

	f1 := createFile(t, tempdir, "hello.txt", []byte("hello.\n"))
	f2 := createFile(t, tempdir, "world.txt", []byte("world.\n"))

	// act
	out, err := exec.Command(build, "-C", tempdir, "add", f1).CombinedOutput()
	if err != nil {
		t.Fatal("add a file failed ", string(out))
	}
	out, err = exec.Command(build, "-C", tempdir, "add", f2).CombinedOutput()
	if err != nil {
		t.Fatal("add a file failed ", string(out))
	}

	// assert
	testlsfiles(t, tempdir, "hello.txt\nworld.txt\n")

}

func TestAddNonExistentFile(t *testing.T) {

	build := buildpath(t)
	tempdir := t.TempDir()

	_, err := exec.Command(build, "init", tempdir).Output()
	if err != nil {
		t.Fatal("init repository failed ", err)
	}

	out, err := exec.Command(build, "-C", tempdir, "add", "/path/to/non/existent/file").CombinedOutput()
	if err == nil {
		t.Fatal("expect error, got nil")
	}

	expectMsg := "stat /path/to/non/existent/file: no such file or directory"
	if string(out) != expectMsg {
		t.Fatalf("expect error message %s, got %s", expectMsg, out)
	}

}

func TestAddUnreadbleFiles(t *testing.T) {

	build := buildpath(t)
	tempdir := t.TempDir()

	_, err := exec.Command(build, "init", tempdir).Output()
	if err != nil {
		t.Fatal("init repository failed ", err)
	}

	f1 := createFile(t, tempdir, "hello.txt", []byte("hello.\n"))
	if err := os.Chmod(f1, 0111); err != nil {
		t.Fatal("chmod failed ", err)
	}

	out, err := exec.Command(build, "-C", tempdir, "add", f1).CombinedOutput()
	if err == nil {
		t.Fatal("expect error, got nil")
	}

	expectMsg := fmt.Sprintf("open %s: permission denied", f1)
	if string(out) != expectMsg {
		t.Fatalf("expect error message %s, got %s", expectMsg, out)
	}

}

func TestAddRelativePath(t *testing.T) {

	build := buildpath(t)

	tempdir := initDir(t, build)

	createFile(t, tempdir, "hello.txt", []byte("hello.\n"))

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	os.Chdir(tempdir)

	add(t, build, ".")

	os.Chdir(wd)

	testlsfiles(t, tempdir, "hello.txt\n")

}

func testlsfiles(t *testing.T, dir string, expect string) {

	t.Helper()

	out, err := exec.Command("git", "-C", dir, "ls-files").CombinedOutput()
	if err != nil {
		t.Fatal(err.Error())
	}

	if string(out) != expect {
		t.Fatalf("unexpected ls-files result. %s", out)
	}

}
