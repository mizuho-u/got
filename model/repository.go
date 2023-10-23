package model

import (
	"io"
	"sort"
	"time"

	"github.com/mizuho-u/got/internal"
	"github.com/mizuho-u/got/model/object"
)

type repository struct {
	objects          []object.Object
	index            *index
	workspace        WorkspaceScanner
	workspaceFiles   map[string]Entry
	head             TreeScanner
	changed          map[string]status
	indexChanges     map[string]status
	workspaceChanges map[string]status
	untracked        []string
}

type WorkspaceOption func(*repository) error

func WithIndex(data io.Reader) WorkspaceOption {

	return func(w *repository) error {

		index, err := newIndex(indexSource(data))
		if err != nil {
			return nil
		}

		w.index = index
		return nil
	}

}

func WithWorkspaceScanner(scanner WorkspaceScanner) WorkspaceOption {

	return func(w *repository) error {
		w.workspace = scanner
		return nil
	}

}

func WithTreeScanner(scanner TreeScanner) WorkspaceOption {

	return func(w *repository) error {
		w.head = scanner
		return nil
	}

}
func NewRepository(options ...WorkspaceOption) (*repository, error) {

	index, err := newIndex()
	if err != nil {
		return nil, err
	}

	ws := &repository{
		objects:          []object.Object{},
		index:            index,
		changed:          map[string]status{},
		indexChanges:     map[string]status{},
		workspaceChanges: map[string]status{},
		untracked:        []string{},
		workspaceFiles:   map[string]Entry{}}

	for _, opt := range options {
		if err := opt(ws); err != nil {
			return nil, err
		}
	}

	return ws, nil

}

func (repo *repository) Untracked() []string {

	// 呼び出しのたびにソートするのは無駄かも
	sort.SliceStable(repo.untracked, func(i, j int) bool {
		return repo.untracked[i] < repo.untracked[j]
	})

	return repo.untracked
}

func (repo *repository) Changed() ([]string, map[string]status) {

	files := internal.Keys(repo.changed)

	sort.SliceStable(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	return files, repo.changed
}

func (repo *repository) IndexChanges() ([]string, map[string]status) {

	files := internal.Keys(repo.indexChanges)

	sort.SliceStable(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	return files, repo.indexChanges
}

func (repo *repository) WorkspaceChanges() ([]string, map[string]status) {

	files := internal.Keys(repo.workspaceChanges)

	sort.SliceStable(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	return files, repo.workspaceChanges
}

func (repo *repository) Commit(parent, author, email, message string, now time.Time) (commitId string, err error) {

	entries := []object.Entry{}

	for _, entry := range repo.index.entries {
		entries = append(entries, object.NewTreeEntry(entry.filename, entry.permission(), entry.oid))
	}

	root, err := object.BuildTree(entries)
	if err != nil {
		return commitId, err
	}

	root.Walk(func(tree object.Object) error {

		repo.objects = append(repo.objects, tree)
		return nil

	})

	a := object.NewAuthor(author, email, now)
	commit, err := object.NewCommit(parent, root.OID(), a.String(), message)
	if err != nil {
		return commitId, err
	}
	repo.objects = append(repo.objects, commit)

	commitId = commit.OID()

	return commitId, err

}

func (repo *repository) Add(scanner WorkspaceScanner) ([]object.Object, error) {

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
		repo.objects = append(repo.objects, blob)

		repo.index.add(NewIndexEntry(f.Name(), blob.OID(), f.Stats()))

	}

}

func (repo *repository) Scan() error {

	if err := repo.scan(); err != nil {
		return err
	}

	repo.detectChanges()

	return nil

}

func (repo *repository) scan() error {

	untrackedSet := map[string]struct{}{}

	for {

		p, err := repo.workspace.Next()
		if err != nil {
			return err
		}

		if p == nil {
			break
		}

		repo.workspaceFiles[p.Name()] = p

		if repo.Index().tracked(p.Name()) {
			continue
		}

		entry := p.Name()
		for _, d := range p.Parents() {

			if !repo.Index().tracked(d) {
				entry = d + "/"
				break
			}
		}

		untrackedSet[entry] = struct{}{}
	}

	for k := range untrackedSet {
		repo.untracked = append(repo.untracked, k)
	}

	return nil
}

const (
	statusNone         status = " "
	statusIndexAdded   status = "A"
	statusFileDeleted  status = "D"
	statusFileModified status = "M"
	statusUnchanged    status = statusNone + statusNone
)

type status string

func (s status) ShortFormat() string {
	return string(s)
}

func (s status) LongFormat() string {
	switch s {
	case statusIndexAdded:
		return "new file"
	case statusFileDeleted:
		return "deleted"
	case statusFileModified:
		return "modified"
	default:
		return ""
	}
}

func (repo *repository) detectChanges() {

	head := map[string]object.Entry{}

	repo.head.Walk(func(name string, entry object.Entry) {

		if entry.IsTree() {
			return
		}

		if !repo.index.trackedFile(name) {
			repo.changed[name] = statusFileDeleted + statusNone
			repo.indexChanges[name] = statusFileDeleted
			return
		}

		head[name] = entry

	})

	for _, e := range repo.index.entries {

		indexStatus := statusNone
		if h, ok := head[e.filename]; !ok {
			indexStatus = statusIndexAdded
			repo.indexChanges[e.filename] = indexStatus
		} else if e.oid != h.OID() || e.permission() != h.Permission() {
			indexStatus = statusFileModified
			repo.indexChanges[e.filename] = indexStatus
		}

		workspaceStatus := statusNone
		if stat, ok := repo.workspaceFiles[e.filename]; !ok {
			workspaceStatus = statusFileDeleted
			repo.workspaceChanges[e.filename] = workspaceStatus
		} else if !repo.index.match(stat) {
			workspaceStatus = statusFileModified
			repo.workspaceChanges[e.filename] = workspaceStatus
		}

		status := indexStatus + workspaceStatus
		if status == statusUnchanged {
			continue
		}

		repo.changed[e.filename] = status
	}

}

func (repo *repository) Objects() []object.Object {
	return repo.objects
}

func (repo *repository) Index() Index {
	return repo.index
}
