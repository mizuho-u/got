package usecase_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func TestFirstCommit(t *testing.T) {

	// arrange
	dir := initDir(t)
	createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	// act
	commitId, err := usecase.Commit(newContext(dir), "First Commit.", time.Unix(1677142145, 0))

	// assert
	if err != nil {
		t.Fatal("failed to commit. ", err)
	}

	head, err := os.ReadFile(filepath.Join(dir, ".git", "HEAD"))
	if err != nil {
		t.Fatal("failed to open HEAD. ", err)
	}

	if commitId != string(head) {
		t.Fatalf("commitId not match to head. head %s commmit %s", string(head), commitId)
	}

}

func TestCommitExcutableFiles(t *testing.T) {

	// arrange
	dir := initDir(t)
	hello := createFile(t, dir, "hello.txt", []byte("Hello world.\n"))

	if err := os.Chmod(hello, 0755); err != nil {
		t.Fatal("failed to chmod test file. ", err)
	}

	// act
	commitId, err := usecase.Commit(newContext(dir), "commit a executable file", time.Unix(1677142145, 0))

	// assert
	if err != nil {
		t.Fatal("failed to commit. ", err)
	}

	head, err := os.ReadFile(filepath.Join(dir, ".git", "HEAD"))
	if err != nil {
		t.Fatal("failed to open HEAD. ", err)
	}

	if commitId != string(head) {
		t.Fatalf("commitId not match to head. head %s commmit %s", string(head), commitId)
	}

}
