package usecase_test

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func initDir(t testing.TB) string {

	t.Helper()

	dir := t.TempDir()

	err := usecase.InitDir(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{}))
	if err != nil {
		t.Fatal(err)
	}

	return dir

}

func add(t *testing.T, dir string, f ...string) {

	t.Helper()

	out := &bytes.Buffer{}
	if err := usecase.Add(newContext(dir, "", "", out, out), f...); err != nil {
		t.Fatal(err)
	}

}

func commit(t *testing.T, dir, email, user, msg string, time time.Time) string {

	t.Helper()

	out := &bytes.Buffer{}
	if err := usecase.Commit(newContext(dir, email, user, out, out), msg, time); err != nil {
		t.Fatal(err)
	}

	return out.String()

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
