package fs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mizuho-u/got/io/database/fs"
)

func TestLockfile(t *testing.T) {

	// arrange
	dir, err := ioutil.TempDir("", "got-test")
	if err != nil {
		t.Fatal("failed to create tempdir", err)
	}
	defer os.RemoveAll(dir)

	// act
	lockfile, err := fs.NewLockfile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal("failed to create lockfile. ", err)
	}

	if err := lockfile.Write([]byte("hello world")); err != nil {
		t.Fatal("failed to write data. ", err)
	}

	if err := lockfile.Commit(); err != nil {
		t.Fatal("failed to commit the lockfile. ", err)
	}

	// assert
	data, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal("failed to read the file. ", err)
	}

	if string(data) != "hello world" {
		t.Fatalf("read data not match. expect %s got %s", "hello world", data)
	}

}

func TestFileAlreadyLocked(t *testing.T) {

	// arrange
	dir, err := ioutil.TempDir("", "got-test")
	if err != nil {
		t.Fatal("failed to create tempdir", err)
	}
	defer os.RemoveAll(dir)

	// act
	// assert
	lockfile, err := fs.NewLockfile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal("failed to create lockfile. ", err)
	}
	defer lockfile.Commit()

	_, err = fs.NewLockfile(filepath.Join(dir, "hello.txt"))
	if err == nil {
		t.Fatal("expected to not get lockfile, but got. ")
	}

}
