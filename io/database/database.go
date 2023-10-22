package database

import (
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/model/object"
)

type Database interface {
	Refs() Refs
	Objects() Objects
	Index() Index
	Close() error
}

type Refs interface {
	HEAD() (string, error)
	UpdateHEAD(commitId string) error
}

type Objects interface {
	Store(objects ...object.Object) error
	Load(oid string) (object.Object, error)
	ScanTree(oid string) model.TreeScanner
}

type Index interface {
	OpenForUpdate() error
	OpenForRead() error
	Update(index model.Index) error
	Read(p []byte) (n int, err error)
	IsNew() bool
}
