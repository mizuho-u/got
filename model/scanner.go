package model

import (
	"io"

	"github.com/mizuho-u/got/model/object"
)

type WorkspaceScanner interface {
	Next() (WorkspaceEntry, error)
}

type WorkspaceEntry interface {
	Name() string
	Size() int64
	Parents() []string
	Stats() *FileStat
	io.ReadSeeker
}

type TreeScanner interface {
	Walk(f func(name string, obj object.Entry))
}
