package model

import (
	"io"
	"time"

	"github.com/mizuho-u/got/model/object"
)

type workspace struct {
	objects []object.Object
	index   *index
}

type WorkspaceOption func(*workspace) error

func WithIndex(data io.Reader) WorkspaceOption {

	return func(w *workspace) error {

		index, err := newIndex(indexSource(data))
		if err != nil {
			return nil
		}

		w.index = index
		return nil
	}

}

func NewWorkspace(options ...WorkspaceOption) (*workspace, error) {

	index, err := newIndex()
	if err != nil {
		return nil, err
	}

	ws := &workspace{objects: []object.Object{}, index: index}

	for _, opt := range options {
		if err := opt(ws); err != nil {
			return nil, err
		}
	}

	return ws, nil

}

func (w *workspace) Commit(parent, author, email, message string, now time.Time, files ...*File) (commitId string, err error) {

	entries := []object.Entry{}

	for _, f := range files {

		blob, err := object.NewBlob(f.Name, f.Data)
		if err != nil {
			return commitId, err
		}
		w.objects = append(w.objects, blob)

		permission := object.RegularFile
		if f.IsExecutable() {
			permission = object.ExecutableFile
		}

		entries = append(entries, object.NewTreeEntry(f.Name, permission, blob.OID()))
	}

	root, err := object.BuildTree(entries)
	if err != nil {
		return commitId, err
	}

	root.Walk(func(tree object.Object) error {

		w.objects = append(w.objects, tree)
		return nil

	})

	a := object.NewAuthor(author, email, now)
	commit, err := object.NewCommit(parent, root.OID(), a.String(), message)
	if err != nil {
		return commitId, err
	}
	commitId = commit.OID()
	w.objects = append(w.objects, commit)

	return commitId, err

}

func (w *workspace) Add(files ...*File) error {

	for _, f := range files {

		blob, err := object.NewBlob(f.Name, f.Data)
		if err != nil {
			return err
		}
		w.objects = append(w.objects, blob)

		w.index.add(NewIndexEntry(f.Name, blob.OID(), f.Stat))

	}

	return nil
}

func (w *workspace) Objects() []object.Object {
	return w.objects
}

func (w *workspace) Index() Index {
	return w.index
}
