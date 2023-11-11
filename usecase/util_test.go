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
	if err := usecase.Add(newContext(dir, "", "", out, out), f); err != nil {
		t.Fatal(err)
	}

}

func commit(t *testing.T, dir, msg string, time time.Time) {

	t.Helper()

	out := &bytes.Buffer{}
	if err := usecase.Commit(newContext(dir, "", "", out, out), msg, time); err != nil {
		t.Fatal(err)
	}

}
