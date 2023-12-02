package usecase_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mizuho-u/got/usecase"
)

func TestInitDir(t *testing.T) {

	// arrange
	dir := t.TempDir()

	out := &bytes.Buffer{}

	// act
	err := usecase.InitDir(newContext(dir, "", "", out, out))
	if err != nil {
		t.Fatal(err)
	}

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

	if out.String() != "Initialized empty Got repository in "+dir {
		t.Errorf("expect outputmsg \"%s\" got %s", "Initialized empty Git repository in "+dir, out.String())
	}

}
