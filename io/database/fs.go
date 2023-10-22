package database

import (
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/io/database/internal/fs"
)

type fsdb struct {
	wsroot  string
	gotroot string
	refs    *fs.Refs
	objects *fs.Objects
	index   *fs.Index
}

func NewFSDB(wsroot, gotroot string) *fsdb {
	return &fsdb{wsroot: wsroot, gotroot: gotroot, refs: fs.NewRefs(gotroot), objects: fs.NewObjects(gotroot), index: fs.NewIndex(gotroot)}
}

func (f *fsdb) Init() error {

	if err := os.MkdirAll(filepath.Join(f.gotroot, "objects"), os.ModeDir|0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(f.gotroot, "refs"), os.ModeDir|0755); err != nil {
		return err
	}

	if err := fs.NewRefs(f.gotroot).UpdateHEAD("ref: refs/heads/main"); err != nil {
		return err
	}

	return nil
}

func (fs *fsdb) Refs() refs {
	return fs.refs
}

func (fs *fsdb) Objects() objects {
	return fs.objects
}

func (fs *fsdb) Index() index {
	return fs.index
}

func (fs *fsdb) Close() error {
	return fs.index.Close()
}
