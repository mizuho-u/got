package usecase_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/mizuho-u/got/usecase"
)

func TestAddSingleFile(t *testing.T) {

	// arrange
	dir := initDir(t)
	f := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	// act
	if code := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f); code != 0 {
		t.Fatal("expect exit code 0, got ", code)
	}

	// assert
	testlsfiles(t, dir, "hello.txt\n")

}

func TestAddMultipleFiles(t *testing.T) {

	// arrange
	dir := initDir(t)
	f1 := createFile(t, dir, "hello.txt", []byte("hello.\n"))
	f2 := createFile(t, dir, "world.txt", []byte("world.\n"))

	// act
	code := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f1, f2)

	// assert
	testExitCode(t, 0, code)
	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestAddFilesFromDirectory(t *testing.T) {

	// arrange
	dir := initDir(t)
	createFile(t, dir, "hello.txt", []byte("hello.\n"))
	createFile(t, dir, "world.txt", []byte("world.\n"))

	// act
	code := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), dir)

	// assert
	testExitCode(t, 0, code)
	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestModifyTheIndex(t *testing.T) {

	// arrange
	dir := initDir(t)
	f1 := createFile(t, dir, "hello.txt", []byte("hello.\n"))
	f2 := createFile(t, dir, "world.txt", []byte("world.\n"))

	// act
	code := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f1)
	testExitCode(t, 0, code)

	code = usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f2)
	testExitCode(t, 0, code)

	// assert
	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestAddNonExistentFile(t *testing.T) {

	dir := initDir(t)

	out := &bytes.Buffer{}
	code := usecase.Add(newContext(dir, out, out), "/path/to/non/existent/file")

	testExitCode(t, 128, code)

}

func TestAddUnreadbleFiles(t *testing.T) {

	dir := initDir(t)
	f1 := createFile(t, dir, "hello.txt", []byte("hello.\n"))

	if err := os.Chmod(f1, 0111); err != nil {
		t.Fatal("chmod failed ", err)
	}

	out := &bytes.Buffer{}
	code := usecase.Add(newContext(dir, out, out), f1)

	testExitCode(t, 128, code)

}

func TestOtherProcessesLockingTheIndex(t *testing.T) {

	dir := initDir(t)
	f1 := createFile(t, dir, ".git/index.lock", []byte(""))

	out := &bytes.Buffer{}
	code := usecase.Add(newContext(dir, out, out), f1)

	testExitCode(t, 128, code)

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
