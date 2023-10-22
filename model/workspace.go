package model

import (
	"io"
	"sort"
	"time"

	"github.com/mizuho-u/got/internal"
	"github.com/mizuho-u/got/model/object"
)

type workspace struct {
	objects     []object.Object
	index       *index
	scanner     WorkspaceScanner
	treeScanner TreeScanner
	changed     map[string]string
	untracked   []string
	stats       map[string]Entry
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

func WithTreeScanner(scanner TreeScanner) WorkspaceOption {

	return func(w *workspace) error {
		w.treeScanner = scanner
		return nil
	}

}
func NewWorkspace(options ...WorkspaceOption) (*workspace, error) {

	index, err := newIndex()
	if err != nil {
		return nil, err
	}

	ws := &workspace{objects: []object.Object{}, index: index, changed: map[string]string{}, untracked: []string{}, stats: map[string]Entry{}}

	for _, opt := range options {
		if err := opt(ws); err != nil {
			return nil, err
		}
	}

	return ws, nil

}

func (w *workspace) Untracked() []string {

	// 呼び出しのたびにソートするのは無駄かも
	sort.SliceStable(w.untracked, func(i, j int) bool {
		return w.untracked[i] < w.untracked[j]
	})

	return w.untracked
}

func (w *workspace) Changed() ([]string, map[string]string) {

	files := internal.Keys(w.changed)

	sort.SliceStable(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	return files, w.changed
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
	w.objects = append(w.objects, commit)

	commitId = commit.OID()

	return commitId, err

}

func (w *workspace) Add(scanner WorkspaceScanner) ([]object.Object, error) {

	objects := []object.Object{}

	for {

		f, err := scanner.Next()
		if err != nil {
			return nil, err
		}
		if f == nil {
			return objects, nil
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		blob, err := object.NewBlob(f.Name(), data)
		if err != nil {
			return nil, err
		}
		w.objects = append(w.objects, blob)

		w.index.add(NewIndexEntry(f.Name(), blob.OID(), f.Stats()))

	}

}

func (w *workspace) Scan() error {

	if err := w.scan(); err != nil {
		return err
	}

	w.detectChanges()

	return nil

}

func (w *workspace) scan() error {

	untrackedSet := map[string]struct{}{}

	for {

		p, err := w.scanner.Next()
		if err != nil {
			return err
		}

		if p == nil {
			break
		}

		w.stats[p.Name()] = p

		if w.Index().tracked(p.Name()) {
			continue
		}

		entry := p.Name()
		for _, d := range p.Parents() {

			if !w.Index().tracked(d) {
				entry = d + "/"
				break
			}
		}

		untrackedSet[entry] = struct{}{}
	}

	for k := range untrackedSet {
		w.untracked = append(w.untracked, k)
	}

	return nil
}

const (
	statusNone         string = " "
	statusIndexAdded   string = "A"
	statusFileDeleted  string = "D"
	statusFileModified string = "M"
)

func (w *workspace) detectChanges() {

	head := map[string]object.Entry{}

	if w.treeScanner != nil {
		w.treeScanner.Walk(func(name string, entry object.Entry) {
			if entry.IsTree() {
				return
			}
			head[name] = entry
		})
	}

	for _, e := range w.index.entries {

		status := ""

		if _, ok := head[e.filename]; !ok {
			status = statusIndexAdded
		} else {
			status = statusNone
		}

		if stat, ok := w.stats[e.filename]; !ok {
			status += statusFileDeleted
		} else if !w.index.match(stat) {
			status += statusFileModified
		} else {
			status += statusNone
		}

		if status == (statusNone + statusNone) {
			continue
		}

		w.changed[e.filename] = status
	}

}

func (w *workspace) Objects() []object.Object {
	return w.objects
}

func (w *workspace) Index() Index {
	return w.index
}
