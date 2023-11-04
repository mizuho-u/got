package e2e

import (
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

var bin = flag.String("build", "", "the build path")

func buildpath(t *testing.T) string {

	t.Helper()

	buildPathAbs, err := filepath.Abs(*bin)
	if err != nil {
		t.Fatal("invalid build path")
	}

	return buildPathAbs

}

func createFile(t testing.TB, dir, name string, data []byte) string {

	t.Helper()

	createDir(t, dir, filepath.Dir(name))

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

	if err := os.MkdirAll(filepath.Join(dir, name), os.ModePerm); err != nil {
		t.Fatal(err)
	}

}

func removeAll(t *testing.T, dir, name string) {

	t.Helper()

	if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
		t.Fatal(err)
	}
}

func initDir(t *testing.T, build string) string {

	t.Helper()

	tempdir := t.TempDir()

	_, err := exec.Command(build, "init", tempdir).Output()
	if err != nil {
		t.Fatal("init repository failed ", err)
	}

	return tempdir
}

func executeCmd(t *testing.T, cmd string) string {

	t.Helper()

	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		t.Fatal("first commit failed ", string(out), err)
	}

	return string(out)
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
