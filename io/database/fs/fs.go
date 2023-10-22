package fs

import (
	"github.com/mizuho-u/got/io/database"
)

type FS struct {
	wsroot  string
	gotroot string
	refs    *refs
	objects *objects
	index   *index
}

func NewFS(wsroot, gotroot string) *FS {
	return &FS{wsroot: wsroot, gotroot: gotroot, refs: NewRefs(gotroot), objects: newObjects(gotroot), index: newIndex(gotroot)}
}

func (fs *FS) Refs() database.Refs {
	return fs.refs
}

func (fs *FS) Objects() database.Objects {
	return fs.objects
}

func (fs *FS) Index() database.Index {
	return fs.index
}

func (fs *FS) Close() error {
	return fs.index.Close()
}
