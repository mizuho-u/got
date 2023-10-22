package e2e

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
