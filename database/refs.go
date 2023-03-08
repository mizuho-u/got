package database

import (
	"io"
	"os"
	"path/filepath"
)

type refs struct {
	gotpath string
}

func NewRefs(gotpath string) *refs {
	return &refs{gotpath}
}

func (r *refs) HEAD() (string, error) {

	head, err := os.Open(filepath.Join(r.gotpath, "HEAD"))
	if err == os.ErrNotExist {
		return "", nil
	}
	defer head.Close()

	commitId, err := io.ReadAll(head)
	if err == os.ErrNotExist {
		return "", err
	}

	return string(commitId), nil
}

func (r *refs) UpdateHEAD(commitId string) error {

	head, err := NewLockfile(filepath.Join(r.gotpath, "HEAD"))
	if err != nil {
		return err
	}
	defer head.Commit()

	err = head.Write([]byte(commitId))
	if err != nil {
		return err
	}

	return nil
}
