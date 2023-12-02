package repository

import (
	"io"

	"github.com/mizuho-u/got/repository/object"
)

type Workspace interface {
	RemoveFile(file string) error
	RemoveDirectory(dir string) error
	CreateDir(dir string) error
	CreateFile(file string) (WorkspaceFile, error)
	Stat(entry string) (WorkspaceFileStat, error)
	ListDir(dir string) ([]WorkspaceFileStat, error)
	Open(f string) (WorkspaceFile, error)
}

type WorkspaceFile interface {
	io.WriteCloser
	io.ReadCloser
	Chmod(p object.Permission) error
	Info() WorkspaceFileStat
}

type WorkspaceFileStat interface {
	Stats() *FileStat
	IsDir() bool
	Name() string
	Path() string
	Size() int64
	Permission() object.Permission
}
