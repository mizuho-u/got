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
	Open(f string) (WorkspaceFile, error)
}

type WorkspaceFile interface {
	io.WriteCloser
	io.ReadCloser
	Permission() object.Permission
	Chmod(p object.Permission) error
}

type WorkspaceFileStat interface {
	IsDir() bool
}
