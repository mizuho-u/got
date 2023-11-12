package fs

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"github.com/mizuho-u/got/repository"
	"github.com/mizuho-u/got/repository/object"
)

type Index struct {
	gotroot  string
	path     string
	file     *os.File
	lockfile *lockfile
}

func NewIndex(gotpath string) *Index {
	return &Index{gotroot: gotpath, path: filepath.Join(gotpath, "index")}
}

func (i *Index) OpenForUpdate() error {

	if err := i.lock(); err != nil {
		return err
	}

	f, err := os.Open(i.path)
	switch {
	case err == nil:
		i.file = f
		return nil
	case errors.Is(err, syscall.ENOENT):
		return nil
	default:
		return err
	}
}

func (i *Index) OpenForRead() error {

	f, err := os.Open(i.path)
	switch {
	case err == nil:
		i.file = f
		return nil
	case errors.Is(err, syscall.ENOENT):
		return nil
	default:
		return err
	}
}

func (i *Index) Update(index repository.Index) error {

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

func (i *Index) Read(p []byte) (n int, err error) {

	if i.file == nil {
		return 0, nil
	}

	return i.file.Read(p)
}

func (i *Index) LoadObject(oid string) (object.Object, error) {
	return load(i.gotroot, oid)
}

func (i *Index) lock() error {

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

func (i *Index) Close() error {

	if i.lockfile == nil {
		return nil
	}

	return i.lockfile.Release()
}

func (i *Index) IsNew() bool {
	return i.file == nil
}
