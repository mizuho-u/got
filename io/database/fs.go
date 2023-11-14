package database

import (
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/io/database/internal/fs"
	"github.com/mizuho-u/got/types"
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

	if err := os.MkdirAll(filepath.Join(f.gotroot, "refs/heads"), os.ModeDir|0755); err != nil {
		return err
	}

	if err := f.refs.UpdateRef("main", ""); err != nil {
		return err
	}

	ref, _ := types.NewBranchName("main")
	if err := f.refs.UpdateHeadRef(ref); err != nil {
		return err
	}

	return nil
}

func (fs *fsdb) Refs() Refs {
	return fs.refs
}

func (fs *fsdb) Objects() Objects {
	return fs.objects
}

func (fs *fsdb) Index() index {
	return fs.index
}

func (fs *fsdb) Close() error {
	return fs.index.Close()
}
