package database

import (
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/model/object"
)

type Database interface {
	Init() error
	Refs() refs
	Objects() objects
	Index() index
	Close() error
}

type refs interface {
	Head() (object.Commit, error)
	UpdateHEAD(commitId string) error
}

type objects interface {
	Store(objects ...object.Object) error
	ScanTree(oid string) model.TreeScanner
	Load(oid string) (object.Object, error)
}

type index interface {
	OpenForUpdate() error
	OpenForRead() error
	Update(index model.Index) error
	Read(p []byte) (n int, err error)
	LoadObject(oid string) (object.Object, error)
	IsNew() bool
}
