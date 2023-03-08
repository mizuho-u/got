package usecase_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mizuho-u/got/usecase"
)

func TestInitDir(t *testing.T) {

	// arrange
	dir, err := ioutil.TempDir("", "got-test")
	if err != nil {
		t.Fatal("failed to create tempdir", err)
	}
	defer os.RemoveAll(dir)

	// act
	if err := usecase.InitDir(dir); err != nil {
		t.Fatal("failed to init dir. ", err)
	}

	// assert
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Error(".got dir not exists.", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".git", "objects")); err != nil {
		t.Error(".got/objects dir not exists.", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".git", "refs")); err != nil {
		t.Error(".got/refs dir not exists.", err)
	}

}

func initDir(t testing.TB) string {

	t.Helper()

	dir := t.TempDir()

	if err := usecase.InitDir(dir); err != nil {
		t.Fatal("failed to init dir. ", err)
	}

	return dir

}

func createFile(t testing.TB, dir, name string, data []byte) string {

	t.Helper()

	file, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal("failed to create test file. ", err)
	}

	t.Cleanup(func() {
		file.Close()
	})

	if _, err := file.Write(data); err != nil {
		t.Fatal("failed to write test file. ", err)
	}

	return file.Name()

}
