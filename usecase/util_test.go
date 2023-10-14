package usecase_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/mizuho-u/got/usecase"
)

func add(t *testing.T, dir, f string) {

	t.Helper()

	out := &bytes.Buffer{}
	if code := usecase.Add(newContext(dir, "", "", out, out), f); code != 0 {
		t.Fatal(out)
	}

}

func commit(t *testing.T, dir, msg string, time time.Time) {

	t.Helper()

	out := &bytes.Buffer{}
	if code := usecase.Commit(newContext(dir, "", "", out, out), msg, time); code != 0 {
		t.Fatal(out)
	}

}
