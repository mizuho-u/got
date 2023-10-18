package model

import "io"

type WorkspaceScanner interface {
	Next() Entry
}

type Entry interface {
	Name() string
	Size() int64
	Parents() []string
	Stats() *FileStat
	io.Reader
}
