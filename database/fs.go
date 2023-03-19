package database

type FS struct {
	root    string
	refs    *refs
	objects *objects
	index   *index
}

func NewFS(root string) *FS {
	return &FS{root: root, refs: NewRefs(root), objects: NewObjects(root), index: newIndex(root)}
}

func (fs *FS) Refs() Refs {
	return fs.refs
}

func (fs *FS) Objects() Objects {
	return fs.objects
}

func (fs *FS) Index() Index {
	return fs.index
}

func (fs *FS) Close() error {
	return fs.index.Close()
}
