package model

import (
	"io"
	"sort"
	"time"

	"github.com/mizuho-u/got/model/object"
)

type workspace struct {
	objects   []object.Object
	index     *index
	scanner   WorkspaceScanner
	changed   map[string]struct{}
	untracked []string
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

func WithWorkspaceScanner(scanner WorkspaceScanner) WorkspaceOption {

	return func(w *workspace) error {
		w.scanner = scanner
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

func (w *workspace) Untracked() []string {
	return w.untracked
}

func (w *workspace) Commit(parent, author, email, message string, now time.Time) (commitId string, err error) {

	entries := []object.Entry{}

	for _, entry := range w.index.entries {
		entries = append(entries, object.NewTreeEntry(entry.filename, entry.permission(), entry.oid))
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

func (w *workspace) Add(f *File) (object.Object, error) {

	blob, err := object.NewBlob(f.Name, f.Data)
	if err != nil {
		return nil, err
	}
	w.objects = append(w.objects, blob)

	w.index.add(NewIndexEntry(f.Name, blob.OID(), f.Stat))

	return blob, err
}

func (w *workspace) Scan() error {

	untrackedSet := map[string]struct{}{}

	for {

		p := w.scanner.Next()
		if p == nil {
			break
		}

		if w.Index().Tracked(p.Name()) {
			continue
		}

		entry := p.Name()
		for _, d := range p.Parents() {

			if !w.Index().Tracked(d) {
				entry = d + "/"
				break
			}
		}

		untrackedSet[entry] = struct{}{}
	}

	untracked := make([]string, 0, len(untrackedSet))
	for k := range untrackedSet {
		untracked = append(untracked, k)
	}

	sort.SliceStable(untracked, func(i, j int) bool {
		return untracked[i] < untracked[j]
	})

	w.untracked = untracked

	return nil

}

func (w *workspace) Objects() []object.Object {
	return w.objects
}

func (w *workspace) Index() Index {
	return w.index
}
