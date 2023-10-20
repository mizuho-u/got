package database

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mizuho-u/got/internal"
	"github.com/mizuho-u/got/model"
)

type FS struct {
	wsroot  string
	gotroot string
	refs    *refs
	objects *objects
	index   *index
}

func NewFS(wsroot, gotroot string) *FS {
	return &FS{wsroot: wsroot, gotroot: gotroot, refs: NewRefs(gotroot), objects: NewObjects(gotroot), index: newIndex(gotroot)}
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

func (fs *FS) Scan(name string) (model.WorkspaceScanner, error) {
	return newFileScanner(fs.wsroot, name, fs.gotroot)
}

type file struct {
	name  string
	size  int64
	stats *model.FileStat
	io.Reader
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Size() int64 {
	return f.size
}

func (f *file) Parents() []string {

	parentsDirs := []string{}
	dir := filepath.Dir(f.name)
	if dir == "." {
		return []string{}
	}

	dirs := strings.Split(filepath.Dir(f.name), "/")
	for i := 1; i <= len(dirs); i++ {
		parentsDirs = append(parentsDirs, filepath.Join(dirs[0:i]...))
	}

	return parentsDirs

}

func (f *file) Stats() *model.FileStat {
	return f.stats
}

type fileScanner struct {
	root   string
	ignore string
	files  internal.Queue[*file]  // rootからのrelpath
	dirs   internal.Queue[string] // fullpath
}

func newFileScanner(root, name, ignore string) (*fileScanner, error) {

	scanner := &fileScanner{root: root, ignore: ignore, files: internal.Queue[*file]{}, dirs: internal.Queue[string]{}}

	info, err := os.Stat(name)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		scanner.dirs = append(scanner.dirs, name)
	} else if err := scanner.enqueueFile(filepath.Dir(name), info); err != nil {
		return nil, err
	}

	return scanner, nil
}

func (fs *fileScanner) Next() (model.Entry, error) {

	if f, err := fs.files.Dequeue(); err == nil {
		return f, nil
	}

	dir, err := fs.dirs.Dequeue()
	if err != nil {
		return nil, nil
	}

	if fs.ignore != "" && strings.HasPrefix(dir, fs.ignore) {
		return fs.Next()
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {

		if entry.IsDir() {
			fs.dirs.Enqueue(filepath.Join(dir, entry.Name()))
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		fs.enqueueFile(dir, info)
	}

	return fs.Next()
}

func (fs *fileScanner) enqueueFile(dir string, info fs.FileInfo) error {

	path, err := filepath.Rel(fs.root, filepath.Join(dir, info.Name()))
	if err != nil {
		return err
	}

	statt, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("failed to get statt")
	}

	reader, err := os.Open(filepath.Join(dir, info.Name()))
	if err != nil {
		return err
	}

	fs.files.Enqueue(&file{name: path, size: info.Size(), stats: model.NewFileStat(statt), Reader: reader})

	return nil
}
