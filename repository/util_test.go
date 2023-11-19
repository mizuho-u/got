package repository_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/mizuho-u/got/internal"
	"github.com/mizuho-u/got/repository"
	"github.com/mizuho-u/got/repository/object"
)

type file struct {
	permission object.Permission
	data       []byte
}

func newTree(t testing.TB, files map[string]*file) (object.Tree, []object.Object) {

	t.Helper()

	entries := []object.TreeEntry{}
	objects := []object.Object{}

	for path, f := range files {

		blob, err := object.NewBlob(path, f.data)
		if err != nil {
			t.Fatal(err)
		}

		objects = append(objects, blob)
		entries = append(entries, object.NewTreeEntry(path, f.permission, blob.OID()))

	}

	tree, err := object.BuildTree(entries)
	if err != nil {
		t.Fatal(err)
	}
	objects = append(objects, tree)

	tree.Walk(func(tree object.Object) error {
		objects = append(objects, tree)
		return nil
	})

	return tree, objects

}

var _ repository.Workspace = &workspace{}

type workspace struct {
	files map[string]repository.WorkspaceFile
	stats map[string]repository.WorkspaceFileStat
}

func newWorkspace() *workspace {

	ws := &workspace{files: map[string]repository.WorkspaceFile{}, stats: map[string]repository.WorkspaceFileStat{}}

	return ws
}

func (ws *workspace) RemoveFile(file string) error {

	delete(ws.files, file)
	delete(ws.stats, file)

	return nil
}

func (ws *workspace) RemoveDirectory(dir string) error {

	delete(ws.stats, dir)

	return nil
}

func (ws *workspace) CreateDir(dir string) error {

	ws.stats[dir] = workspaceFileStat(true)

	return nil
}

func (ws *workspace) CreateFile(file string) (repository.WorkspaceFile, error) {

	ws.files[file] = &workspaceFile{object.RegularFile, []byte{}}
	ws.stats[file] = workspaceFileStat(false)

	return ws.files[file], nil
}

func (ws *workspace) Stat(entry string) (repository.WorkspaceFileStat, error) {

	if stat, ok := ws.stats[entry]; ok {

		return stat, nil

	} else {

		return nil, fmt.Errorf("%s not found", entry)

	}
}

func (ws *workspace) Open(f string) (repository.WorkspaceFile, error) {

	if file, ok := ws.files[f]; ok {

		return file, nil

	} else {

		return nil, fmt.Errorf("%s not found", file)

	}
}

func (ws *workspace) add(path string, f *file) {

	ws.files[path] = &workspaceFile{f.permission, f.data}
	ws.stats[path] = workspaceFileStat(false)

	for _, d := range internal.ParentDirs(path) {
		ws.stats[d] = workspaceFileStat(true)
	}
}

func (ws *workspace) addRange(files map[string]*file) {

	for path, f := range files {
		ws.add(path, f)
	}
}

func (ws *workspace) equals(t testing.TB, files map[string]*file) {

	t.Helper()

	for path, expect := range files {

		got, ok := ws.files[path]
		if !ok {
			t.Errorf("%s not found", expect)
		}

		data, err := io.ReadAll(got)
		if err != nil {
			t.Error(err)
		}

		if len(data) != len(expect.data) {
			t.Errorf("expect %s data len %d, got %d", path, len(expect.data), len(data))
		}

		if string(data) != string(expect.data) {
			t.Errorf("expect %s contents %s, got %s", path, expect.data, data)
		}

		if got.Permission() != expect.permission {
			t.Errorf("expect %s permission %s, got %s", path, expect.permission, got.Permission())
		}

	}

	expectDirs := internal.NewSet[string]()
	for path := range files {

		p := internal.NewSetFromArray(internal.ParentDirs(path))
		expectDirs.Merge(p)
	}

	for path, entry := range ws.stats {

		if !entry.IsDir() {
			continue
		}

		if !expectDirs.Has(path) {
			t.Errorf("unexpect directory exists: %s", path)
		}

	}

}

var _ repository.WorkspaceFile = &workspaceFile{}

type workspaceFile struct {
	permission object.Permission
	data       []byte
}

func (f *workspaceFile) Chmod(p object.Permission) error {
	f.permission = p
	return nil
}

func (f *workspaceFile) Write(data []byte) (int, error) {
	f.data = data
	return 0, nil
}

func (f *workspaceFile) Close() error {
	return nil
}

func (f *workspaceFile) Read(p []byte) (n int, err error) {

	if len(f.data) == 0 {
		return 0, io.EOF
	}

	for i := 0; i < len(p) && i < len(f.data); i++ {
		p[i] = f.data[i]
		n++
	}

	err = io.EOF

	return
}

func (f *workspaceFile) Permission() object.Permission {
	return f.permission
}

var _ repository.WorkspaceFileStat = workspaceFileStat(true)

type workspaceFileStat bool

func (s workspaceFileStat) IsDir() bool {
	return bool(s)
}

func (s workspaceFileStat) Stats() *repository.FileStat {
	return nil
}

var _ repository.Database = &database{}

type database struct {
	objects map[string]object.Object
}

func newDatabase() *database {
	return &database{map[string]object.Object{}}
}

func (db *database) Load(oid string) (object.Object, error) {

	if o, ok := db.objects[oid]; ok {
		return o, nil
	} else {
		return nil, fmt.Errorf("object %s not found", oid)
	}

}

func (db *database) store(objects ...object.Object) {

	for _, o := range objects {
		db.objects[o.OID()] = o
	}

}

var _ repository.IndexWriter = &index{}

type index struct {
	entries map[string]*repository.IndexEntry
}

func newIndex() *index {
	return &index{map[string]*repository.IndexEntry{}}
}

func (i *index) Add(entries ...*repository.IndexEntry) {

	for _, entry := range entries {
		i.entries[entry.Name()] = entry
	}

}

func (i *index) Delete(entry string) {
	delete(i.entries, entry)
}
