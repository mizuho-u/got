package workspace

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mizuho-u/got/internal"
	"github.com/mizuho-u/got/io/workspace/internal/stat"

	"github.com/mizuho-u/got/repository"
	"github.com/mizuho-u/got/repository/object"
)

type fileScanner struct {
	root   string
	ignore string
	files  internal.Queue[*file]  // rootからのrelpath
	dirs   internal.Queue[string] // fullpath
}

// Scan nameをスキャンしてrootDirからの相対パスを取得するfileScannerを生成する
func Scan(rootDir, name, ignore string) (*fileScanner, error) {

	scanner := &fileScanner{root: rootDir, ignore: ignore, files: internal.Queue[*file]{}, dirs: internal.Queue[string]{}}

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

// Next エントリをひとつ返す。最後はnil
func (fs *fileScanner) Next() (repository.WorkspaceEntry, error) {

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

	fs.files.Enqueue(&file{name: path, size: info.Size(), stats: newFileStat(statt), ReadSeeker: reader})

	return nil
}

type file struct {
	name  string
	size  int64
	stats *repository.FileStat
	io.ReadSeeker
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

func (f *file) Stats() *repository.FileStat {
	return f.stats
}

var _ repository.Workspace = &workspace{}

type workspace struct {
	root string
}

func New(root string) *workspace {
	return &workspace{root: root}
}

func (ws *workspace) abs(path string) string {
	return filepath.Join(ws.root, path)
}

func (ws *workspace) rel(path string) (string, error) {
	return filepath.Rel(ws.root, path)
}

func (ws *workspace) RemoveFile(file string) error {
	return os.Remove(ws.abs(file))
}

func (ws *workspace) RemoveDirectory(dir string) error {
	return os.Remove(ws.abs(dir))
}

func (ws *workspace) CreateDir(dir string) error {
	return os.Mkdir(ws.abs(dir), 0755)
}

func (ws *workspace) CreateFile(file string) (repository.WorkspaceFile, error) {

	f, err := os.OpenFile(ws.abs(file), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0664)
	if err != nil {
		return nil, err
	}

	return newWorkspaceFile(ws.root, f)
}

func (ws *workspace) Stat(entry string) (repository.WorkspaceFileStat, error) {

	stat, err := os.Stat(ws.abs(entry))
	if err != nil {
		return nil, err
	}

	return newWorkspaceFileStat(ws.root, stat)
}

func (ws *workspace) ListDir(dir string) ([]repository.WorkspaceFileStat, error) {

	entries, err := os.ReadDir(ws.abs(dir))
	if err != nil {
		return nil, err
	}

	stats := []repository.WorkspaceFileStat{}
	for _, e := range entries {

		info, err := e.Info()
		if err != nil {
			return nil, err
		}

		stat, err := newWorkspaceFileStat(filepath.Join(dir, info.Name()), info)

		if err != nil {
			return nil, err
		}

		stats = append(stats, stat)
	}

	return stats, nil

}

func (ws *workspace) Walk(dir string, walkFunc func(file repository.WorkspaceFile) error) error {

	return filepath.WalkDir(ws.abs(dir), func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		rel, err := ws.rel(path)
		if err != nil {
			return err
		}

		file, err := ws.Open(rel)
		if err != nil {
			return err
		}

		return walkFunc(file)

	})

}

func (ws *workspace) Open(file string) (repository.WorkspaceFile, error) {

	perm := os.O_RDWR
	if _, err := os.Stat(ws.abs(file)); err != nil {
		perm |= os.O_CREATE | os.O_EXCL
		os.MkdirAll(filepath.Dir(ws.abs(file)), 0755)
	}

	f, err := os.OpenFile(ws.abs(file), perm, 0664)
	if err != nil {
		return nil, err
	}

	return newWorkspaceFile(file, f)
}

func statToPermission(stat fs.FileInfo) object.Permission {

	if stat.IsDir() {
		return object.Directory
	}

	if (stat.Mode() & 0111) == 0111 {
		return object.ExecutableFile
	}

	return object.RegularFile

}

var _ repository.WorkspaceFile = &workspaceFile{}

type workspaceFile struct {
	*os.File
	stat repository.WorkspaceFileStat
}

func newWorkspaceFile(path string, f *os.File) (*workspaceFile, error) {

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	stat, err := newWorkspaceFileStat(path, info)
	if err != nil {
		return nil, err
	}

	return &workspaceFile{f, stat}, nil

}

func (f *workspaceFile) Chmod(p object.Permission) error {

	var mode fs.FileMode
	switch p {
	case object.ExecutableFile:
		mode = 0755
	case object.RegularFile:
		mode = 0664
	default:
		return fmt.Errorf("invalid permision %s", p)
	}

	return f.File.Chmod(mode)
}

func (f *workspaceFile) Info() repository.WorkspaceFileStat {
	return f.stat
}

var _ repository.WorkspaceFileStat = &workspaceFileStat{}

type workspaceFileStat struct {
	path string
	fs.FileInfo
	stats *repository.FileStat
}

func newWorkspaceFileStat(path string, info fs.FileInfo) (*workspaceFileStat, error) {

	statt, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, errors.New("failed to get statt")
	}

	return &workspaceFileStat{path, info, newFileStat(statt)}, nil

}

func (stat *workspaceFileStat) IsDir() bool {
	return stat.FileInfo.IsDir()
}

func (stat *workspaceFileStat) Stats() *repository.FileStat {
	return stat.stats
}

func (stat *workspaceFileStat) Name() string {
	return stat.FileInfo.Name()
}

func (stat *workspaceFileStat) Path() string {
	return stat.path
}

func (stat *workspaceFileStat) Size() int64 {
	return stat.FileInfo.Size()
}

func (stat *workspaceFileStat) Permission() object.Permission {
	return stat.stats.Permission()
}

func newFileStat(s *syscall.Stat_t) *repository.FileStat {

	cspec := stat.Ctimespec(s)
	mspec := stat.Mtimespec(s)

	return repository.NewFileStat(
		uint32(cspec.Sec),
		uint32(cspec.Nsec),
		uint32(mspec.Sec),
		uint32(mspec.Nsec),
		uint32(s.Dev),
		uint32(s.Ino),
		uint32(s.Mode),
		uint32(s.Uid),
		uint32(s.Gid),
		uint32(s.Size),
	)

}
