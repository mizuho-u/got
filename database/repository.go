package database

import (
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/model/object"
)

type Repository interface {
	Refs() Refs
	Objects() Objects
	Index() Index
	Scan() model.WorkspaceScanner
	Close() error
}

type Refs interface {
	HEAD() (string, error)
	UpdateHEAD(commitId string) error
}

type Objects interface {
	Store(o object.Object) error
}

type Index interface {
	OpenForUpdate() error
	OpenForRead() error
	Update(index model.Index) error
	Read(p []byte) (n int, err error)
	IsNew() bool
}
