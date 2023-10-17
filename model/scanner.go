package model

type WorkspaceScanner interface {
	Next() Entry
}

type Entry interface {
	Name() string
	Size() int64
	Parents() []string
}
