package database

import (
	"io"
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

func (fs *FS) Scan() model.WorkspaceScanner {
	return newFileScanner(fs.wsroot, fs.gotroot)
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

func newFileScanner(dir, ignore string) *fileScanner {
	return &fileScanner{root: dir, ignore: ignore, files: internal.Queue[*file]{}, dirs: internal.Queue[string]{dir}}
}

func (fs *fileScanner) Next() model.Entry {

	if f, err := fs.files.Dequeue(); err == nil {
		return f
	}

	dir, err := fs.dirs.Dequeue()
	if err != nil {
		return nil
	}

	if fs.ignore != "" && strings.HasPrefix(dir, fs.ignore) {
		return fs.Next()
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {

		if entry.IsDir() {
			fs.dirs.Enqueue(filepath.Join(dir, entry.Name()))
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil
		}

		statt, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return nil
		}

		path, err := filepath.Rel(fs.root, filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil
		}

		reader, err := os.Open(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil
		}

		fs.files.Enqueue(&file{name: path, size: info.Size(), stats: model.NewFileStat(statt), Reader: reader})
	}

	return fs.Next()
}
