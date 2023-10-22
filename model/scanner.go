package model

import (
	"io"

	"github.com/mizuho-u/got/model/object"
)

type WorkspaceScanner interface {
	Next() (Entry, error)
}

type Entry interface {
	Name() string
	Size() int64
	Parents() []string
	Stats() *FileStat
	io.Reader
}

type TreeScanner interface {
	Walk(f func(name string, obj object.Entry))
}
