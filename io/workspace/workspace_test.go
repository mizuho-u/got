package workspace_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mizuho-u/got/io/workspace"
	"github.com/mizuho-u/got/repository/object"
)

func TestDirectory(t *testing.T) {

	root := t.TempDir()

	ws := workspace.New(root)

	if err := ws.CreateDir("test"); err != nil {
		t.Fatal(err)
	}

	stat, err := os.Stat(filepath.Join(root, "test"))
	if err != nil {
		t.Fatal(err)
	}

	if !stat.IsDir() {
		t.Fatal("test is not directory")
	}

	if stat, err := ws.Stat("test"); err != nil || !stat.IsDir() {
		t.Fatalf("test is not directory %s", err)
	}

	if err := ws.RemoveDirectory("test"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, "test")); err == nil {
		t.Fatal("test still exists")
	}

	if _, err := ws.Stat(filepath.Join(root, "test")); err == nil {
		t.Fatal("test still exists")
	}

}

func TestFile(t *testing.T) {

	root := t.TempDir()

	ws := workspace.New(root)

	if _, err := ws.CreateFile("test/file.txt"); err == nil {
		t.Fatal("unexpected file created")
	}

	if err := ws.CreateDir("test"); err != nil {
		t.Fatal(err)
	}

	if _, err := ws.CreateFile("test/file.txt"); err != nil {
		t.Fatal(err)
	}

	if stat, err := ws.Stat("test/file.txt"); err != nil || stat.IsDir() {
		t.Fatal(err)
	}

	if err := ws.RemoveFile("test/file.txt"); err != nil {
		t.Fatal(err)
	}

}

func TestEmptyDirecotryRemovable(t *testing.T) {

	root := t.TempDir()

	ws := workspace.New(root)

	if err := ws.CreateDir("test"); err != nil {
		t.Fatal(err)
	}

	if _, err := ws.CreateFile("test/file.txt"); err != nil {
		t.Fatal(err)
	}

	if err := ws.RemoveDirectory("test"); err == nil {
		t.Fatal("directory removed")
	}

	if err := ws.RemoveFile("test/file.txt"); err != nil {
		t.Fatal(err)
	}

	if err := ws.RemoveDirectory("test"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat("test"); err == nil {
		t.Fatal("test still exists")
	}

}

func TestWriteFile(t *testing.T) {

	root := t.TempDir()

	_, err := os.Create(filepath.Join(root, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}

	ws := workspace.New(root)

	f, err := ws.Open("file.txt")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = f.Write([]byte("abc\n")); err != nil {
		t.Fatal(err)
	}

	modified, err := os.Open(filepath.Join(root, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(modified)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "abc\n" {
		t.Fatalf("expect %s, got %s", "abc\n", data)
	}

}

func TestChmodFile(t *testing.T) {

	root := t.TempDir()

	f, err := os.Create(filepath.Join(root, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}

	if stat, err := f.Stat(); err != nil && stat.Mode()&0111 == 0111 {
		t.Fatal("file.txt is executable")
	}

	ws := workspace.New(root)

	opened, err := ws.Open("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()

	if err := opened.Chmod(object.ExecutableFile); err != nil {
		t.Fatal(err)
	}

	if stat, err := f.Stat(); err != nil && stat.Mode()&0111 != 0111 {
		t.Fatal("file.txt is not executable")
	}

}
