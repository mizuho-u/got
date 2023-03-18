package usecase_test

import (
	"bytes"
	"errors"
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
	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f); err != nil {
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
	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f1, f2); err != nil {
		t.Fatal(err)
	}

	// assert
	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestAddFilesFromDirectory(t *testing.T) {

	// arrange
	dir := initDir(t)
	createFile(t, dir, "hello.txt", []byte("hello.\n"))
	createFile(t, dir, "world.txt", []byte("world.\n"))

	// act
	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), dir); err != nil {
		t.Fatal(err)
	}

	// assert
	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestModifyTheIndex(t *testing.T) {

	// arrange
	dir := initDir(t)
	f1 := createFile(t, dir, "hello.txt", []byte("hello.\n"))
	f2 := createFile(t, dir, "world.txt", []byte("world.\n"))

	// act
	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f1); err != nil {
		t.Fatal(err)
	}
	if err := usecase.Add(newContext(dir, &bytes.Buffer{}, &bytes.Buffer{}), f2); err != nil {
		t.Fatal(err)
	}

	// assert
	testlsfiles(t, dir, "hello.txt\nworld.txt\n")

}

func TestAddNonExistentFile(t *testing.T) {

	dir := initDir(t)

	out := &bytes.Buffer{}
	err := usecase.Add(newContext(dir, out, out), "/path/to/non/existent/file")

	if !errors.Is(err, usecase.ErrMissingFile) {
		t.Fatalf("expect error %s got %s", os.ErrNotExist, err)
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
