package repository_test

import (
	"fmt"
	"io"
	"path/filepath"
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
	dirs  map[string]internal.Set[repository.WorkspaceFileStat]
}

func newWorkspace() *workspace {

	ws := &workspace{
		files: map[string]repository.WorkspaceFile{},
		stats: map[string]repository.WorkspaceFileStat{},
		dirs:  make(map[string]internal.Set[repository.WorkspaceFileStat])}

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

	ws.stats[dir] = newWorkspaceFileStat(dir, 0, object.Directory, true)
	return nil
}

func (ws *workspace) CreateFile(file string) (repository.WorkspaceFile, error) {

	ws.files[file] = newWorkspaceFile(file, object.RegularFile, []byte{})
	ws.stats[file] = ws.files[file].Info()

	return ws.files[file], nil
}

func (ws *workspace) Stat(entry string) (repository.WorkspaceFileStat, error) {

	if stat, ok := ws.stats[entry]; ok {

		return stat, nil

	} else {

		return nil, fmt.Errorf("%s not found", entry)

	}
}

func (ws *workspace) ListDir(dir string) ([]repository.WorkspaceFileStat, error) {

	stats, ok := ws.dirs[dir]
	if !ok {
		return nil, fmt.Errorf("dir %s not found", dir)
	}

	return stats.Iter(), nil
}

func (ws *workspace) Open(f string) (repository.WorkspaceFile, error) {

	if file, ok := ws.files[f]; ok {

		return file, nil

	} else {

		return nil, fmt.Errorf("%s not found", file)

	}
}

func (ws *workspace) add(path string, f *file) {

	ws.files[path] = newWorkspaceFile(path, f.permission, f.data)
	ws.stats[path] = ws.files[path].Info()

	ws.setDirs(filepath.Dir(path), ws.stats[path])

	parents := internal.ParentDirs(path)
	for i, d := range parents {
		ws.stats[d] = newWorkspaceFileStat(d, 0, object.Directory, true)
		if i >= 1 {
			ws.setDirs(parents[i-1], ws.stats[d])
		}
	}

}

func (ws *workspace) setDirs(path string, stat repository.WorkspaceFileStat) {

	dir, ok := ws.dirs[path]
	if !ok {
		ws.dirs[path] = internal.NewSet[repository.WorkspaceFileStat]()
		dir = ws.dirs[path]
	}

	dir.Set(stat)

}

func (ws *workspace) addRange(files map[string]*file) {

	for path, f := range files {
		ws.add(path, f)
	}
}

func (ws *workspace) modify(files map[string]*file) {

	for path, f := range files {
		ws.add(path, f)
		stat := ws.stats[path]
		if stat, ok := stat.(*workspaceFileStat); ok {

			stat.stat = repository.NewFileStat(1, 1, 1, 1, 0, 0, permissionToMode(f.permission), 0, 0, uint32(len(f.data)))
		}

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

		if got.Info().Permission() != expect.permission {
			t.Errorf("expect %s permission %s, got %s", path, expect.permission, got.Info().Permission())
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

func newWorkspaceFile(path string, permission object.Permission, data []byte) *workspaceFile {
	return &workspaceFile{data, newWorkspaceFileStat(path, int64(len(data)), permission, permission == object.Directory)}
}

type workspaceFile struct {
	data []byte
	stat *workspaceFileStat
}

func (f *workspaceFile) Chmod(p object.Permission) error {
	f.stat.permission = p
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

func (f *workspaceFile) Info() repository.WorkspaceFileStat {
	return f.stat
}

var _ repository.WorkspaceFileStat = &workspaceFileStat{}

type workspaceFileStat struct {
	path       string
	len        int64
	permission object.Permission
	isDir      bool
	stat       *repository.FileStat
}

func newWorkspaceFileStat(path string, len int64, permission object.Permission, isDir bool) *workspaceFileStat {

	stat := repository.NewFileStat(0, 0, 0, 0, 0, 0, permissionToMode(permission), 0, 0, uint32(len))

	return &workspaceFileStat{path, len, permission, isDir, stat}

}

func permissionToMode(permission object.Permission) uint32 {

	switch permission {
	case object.Directory:
		return modeDir
	case object.ExecutableFile:
		return modeFileExecutable
	default:
		return modeFileRegurar
	}

}

func (s *workspaceFileStat) IsDir() bool {
	return s.isDir
}

const (
	modeFileRegurar    uint32 = 33188
	modeFileExecutable uint32 = 33261
	modeDir            uint32 = 16877
)

func (s *workspaceFileStat) Stats() *repository.FileStat {
	return s.stat
}

func (s *workspaceFileStat) Name() string {
	return filepath.Base(s.path)
}

func (s *workspaceFileStat) Path() string {
	return s.path
}

func (s *workspaceFileStat) Size() int64 {
	return s.len
}

func (s *workspaceFileStat) Permission() object.Permission {
	return s.permission
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
