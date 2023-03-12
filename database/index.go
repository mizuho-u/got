package database

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"github.com/mizuho-u/got/model"
)

type index struct {
	path     string
	file     *os.File
	lockfile *lockfile
}

func OpenIndexForUpdate(gotpath string) (*index, error) {

	i := &index{path: filepath.Join(gotpath, "index")}

	if err := i.lock(); err != nil {
		return nil, err
	}

	f, err := os.Open(i.path)
	switch {
	case err == nil:
		i.file = f
		return i, nil
	case errors.Is(err, syscall.ENOENT):
		return i, nil
	default:
		return nil, err
	}
}

func OpenIndexForRead(gotpath string) (*index, error) {

	i := &index{path: filepath.Join(gotpath, "index")}

	f, err := os.Open(i.path)
	switch {
	case err == nil:
		i.file = f
		return i, nil
	case errors.Is(err, syscall.ENOENT):
		return i, nil
	default:
		return nil, err
	}
}

func (i *index) Update(index model.Index) error {

	content, err := index.Serialize()
	if err != nil {
		i.lockfile.Release()
		return err
	}

	err = i.lockfile.Write(content)
	if err != nil {
		i.lockfile.Release()
		return err
	}

	return i.lockfile.Commit()
}

func (i *index) Read(p []byte) (n int, err error) {

	if i.file == nil {
		return 0, nil
	}

	return i.file.Read(p)
}

func (i *index) lock() error {

	if i.lockfile != nil {
		return nil
	}

	lock, err := NewLockfile(i.path)
	if err != nil {
		return err
	}

	i.lockfile = lock

	return nil
}

func (i *index) Close() error {

	if i.lockfile == nil {
		return nil
	}

	return i.lockfile.Release()
}

func (i *index) IsNew() bool {
	return i.file == nil
}
