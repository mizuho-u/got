package usecase_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func TestShowHead(t *testing.T) {

	// arrange
	dir := initDir(t)
	createFile(t, dir, "hello.txt", []byte("Hello world 1.\n"))
	createFile(t, dir, "a/hello2.txt", []byte("Hello world 2.\n"))
	createFile(t, dir, "a/b/c/hello3.txt", []byte("Hello world 3.\n"))

	ctx := newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{})
	if err := usecase.Add(ctx, ctx.WorkspaceRoot()); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	if err := usecase.Commit(newContext(dir, "Mizuho Ueda", "mi_ueda@u-m.dev", out, out), "commit\n", time.Unix(1694356071, 0)); err != nil {
		t.Fatal("expect exit code 0, got ", err)
	}

	if err := usecase.ShowHead(newContext(dir, "", "", &bytes.Buffer{}, &bytes.Buffer{})); err != nil {
		t.Fatal(err)
	}

}
