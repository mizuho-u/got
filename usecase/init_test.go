package usecase_test

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func TestInitDir(t *testing.T) {

	// arrange
	dir, err := ioutil.TempDir("", "got-test")
	if err != nil {
		t.Fatal("failed to create tempdir", err)
	}
	defer os.RemoveAll(dir)

	out := &bytes.Buffer{}

	// act
	code := usecase.InitDir(newContext(dir, "", "", out, out))
	testExitCode(t, 0, code)

	// assert
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Error(".git dir not exists.", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".git", "objects")); err != nil {
		t.Error(".git/objects dir not exists.", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".git", "refs")); err != nil {
		t.Error(".git/refs dir not exists.", err)
	}

	if out.String() != "Initialized empty Jit repository in "+dir {
		t.Errorf("expect outputmsg \"%s\" got %s", "Initialized empty Jit repository in "+dir, out.String())
	}

}

func initDir(t testing.TB) string {

	t.Helper()

	dir := t.TempDir()

	code := usecase.InitDir(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}))
	testExitCode(t, 0, code)

	return dir

}

func createFile(t testing.TB, dir, name string, data []byte) string {

	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), os.ModePerm); err != nil {
		t.Fatal(err)
	}

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

func createDir(t testing.TB, dir, name string) {

	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), os.ModePerm); err != nil {
		t.Fatal(err)
	}

}

func modifyFileMode(t testing.TB, dir, name string, mode fs.FileMode) {

	t.Helper()

	if err := os.Chmod(filepath.Join(dir, name), mode); err != nil {
		t.Fatal(err)
	}

}

func modifyFileTime(t testing.TB, dir, name string, atime, mtime time.Time) {

	t.Helper()

	if err := os.Chtimes(filepath.Join(dir, name), atime, mtime); err != nil {
		t.Fatal(err)
	}

}

func removeAll(t testing.TB, dir, name string) {

	t.Helper()

	if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
		t.Fatal(err)
	}

}

func newContext(dir, username, email string, out, outErr io.Writer) usecase.GotContext {
	return usecase.NewContext(context.Background(), dir, ".git", username, email, out, outErr)
}

func testExitCode(t testing.TB, expect, got usecase.ExitCode) {

	if expect != got {
		t.Fatalf("expect exit code %d, got %d", expect, got)
	}

}
