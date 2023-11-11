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
	err := usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), f)
	if err != nil {
		t.Fatal(err)
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
	err := usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), f1, f2)

	// assert
	if err != nil {
		t.Fatal(err)
	}
	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestAddFilesFromDirectory(t *testing.T) {

	// arrange
	dir := initDir(t)
	createFile(t, dir, "hello.txt", []byte("hello.\n"))
	createFile(t, dir, "world.txt", []byte("world.\n"))

	// act
	err := usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), dir)

	// assert
	if err != nil {
		t.Fatal(err)
	}

	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestModifyTheIndex(t *testing.T) {

	// arrange
	dir := initDir(t)
	f1 := createFile(t, dir, "hello.txt", []byte("hello.\n"))
	f2 := createFile(t, dir, "world.txt", []byte("world.\n"))

	// act
	err := usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), f1)
	if err != nil {
		t.Fatal(err)
	}

	err = usecase.Add(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}), f2)
	if err != nil {
		t.Fatal(err)
	}

	// assert
	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestAddNonExistentFile(t *testing.T) {

	dir := initDir(t)

	out := &bytes.Buffer{}
	err := usecase.Add(newContext(dir, "", "", out, out), "/path/to/non/existent/file")
	if err == nil {
		t.Fatalf("expect err %s got nil", err)
	}

}

func TestAddUnreadbleFiles(t *testing.T) {

	dir := initDir(t)
	f1 := createFile(t, dir, "hello.txt", []byte("hello.\n"))

	if err := os.Chmod(f1, 0111); err != nil {
		t.Fatal("chmod failed ", err)
	}

	out := &bytes.Buffer{}
	err := usecase.Add(newContext(dir, "", "", out, out), f1)
	if err == nil {
		t.Fatalf("expect err %s got nil", err)
	}

}

func TestOtherProcessesLockingTheIndex(t *testing.T) {

	dir := initDir(t)
	f1 := createFile(t, dir, ".git/index.lock", []byte(""))

	out := &bytes.Buffer{}
	err := usecase.Add(newContext(dir, "", "", out, out), f1)
	if err == nil {
		t.Fatalf("expect err %s got nil", err)
	}

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
