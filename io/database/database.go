package database

import (
	"github.com/mizuho-u/got/repository"
	"github.com/mizuho-u/got/repository/object"
	"github.com/mizuho-u/got/types"
)

type Database interface {
	Init() error
	Refs() Refs
	Objects() Objects
	Index() index
	Close() error
}

type Refs interface {
	Head() (object.Commit, error)
	UpdateHeadCommit(commitId string) error
	UpdateHeadRef(branchName types.BranchName) error
	CreateBranch(branchName types.BranchName, oid string) error
	Ref(branchName string) (object.Commit, error)
}

type Objects interface {
	Store(objects ...object.Object) error
	ScanTree(oid string) repository.TreeScanner
	Load(oid string) (object.Object, error)
	LoadPrefix(prefix string) ([]object.Object, error)
	LoadCommit(oid string) (object.Commit, error)
}

type index interface {
	OpenForUpdate() error
	OpenForRead() error
	Update(index repository.Index) error
	Read(p []byte) (n int, err error)
	LoadObject(oid string) (object.Object, error)
	IsNew() bool
}
