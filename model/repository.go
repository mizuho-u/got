package model

import (
	"fmt"
	"io"
	"path/filepath"
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
	headScanner      TreeScanner
	head             map[string]object.Entry
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
		w.headScanner = scanner
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
		head:             map[string]object.Entry{},
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

const (
	nullPath string = "/dev/null"
	nullOID  string = "0000000000000000000000000000000000000000"
)

type diff interface {
	PathLine() string
	ModeLine() string
	IndexLine() string
	FileLine() string
}

type diffModified struct {
	AOID  string
	AMode string
	APath string
	BOID  string
	BMode string
	BPath string
}

func (diff *diffModified) PathLine() string {
	return fmt.Sprintf("diff --git %s %s\n", diff.APath, diff.BPath)
}

func (diff *diffModified) ModeLine() string {

	l := ""

	if diff.AMode != diff.BMode {
		l += fmt.Sprintf("old mode %s\n", diff.AMode)
		l += fmt.Sprintf("new mode %s\n", diff.BMode)
	}

	return l
}

func (diff *diffModified) IndexLine() string {

	if diff.AOID == diff.BOID {
		return ""
	}

	if diff.AMode != diff.BMode {
		return fmt.Sprintf("index %s..%s\n", diff.AOID, diff.BOID)
	}

	return fmt.Sprintf("index %s..%s %s\n", diff.AOID, diff.BOID, diff.AMode)
}

func (diff *diffModified) FileLine() string {

	if diff.AOID == diff.BOID {
		return ""
	}

	l := fmt.Sprintf("--- %s\n", diff.APath)
	l += fmt.Sprintf("+++ %s\n", diff.BPath)

	return l

}

type diffDeleted struct {
	AOID  string
	AMode string
	APath string
	BOID  string
	BMode string
	BPath string
}

func (diff *diffDeleted) PathLine() string {
	return fmt.Sprintf("diff --git %s %s\n", diff.APath, diff.BPath)
}

func (diff *diffDeleted) ModeLine() string {
	return fmt.Sprintf("deleted file mode %s\n", diff.AMode)
}

func (diff *diffDeleted) IndexLine() string {
	return fmt.Sprintf("index %s..%s\n", diff.AOID, diff.BOID)
}

func (diff *diffDeleted) FileLine() string {

	l := fmt.Sprintf("--- %s\n", diff.APath)
	l += fmt.Sprintf("+++ %s\n", nullPath)

	return l

}

type diffAdded struct {
	AOID  string
	AMode string
	APath string
	BOID  string
	BMode string
	BPath string
}

func (diff *diffAdded) PathLine() string {
	return fmt.Sprintf("diff --git %s %s\n", diff.APath, diff.BPath)
}

func (diff *diffAdded) ModeLine() string {
	return fmt.Sprintf("new file mode %s\n", diff.BMode)
}

func (diff *diffAdded) IndexLine() string {
	return fmt.Sprintf("index %s..%s\n", diff.AOID, diff.BOID)
}

func (diff *diffAdded) FileLine() string {

	l := fmt.Sprintf("--- %s\n", nullPath)
	l += fmt.Sprintf("+++ %s\n", diff.BPath)

	return l

}

func (repo *repository) Diff(staged bool) ([]diff, error) {

	if staged {
		return repo.diffStaged()
	}

	diffs := []diff{}

	files, types := repo.WorkspaceChanges()

	for _, path := range files {

		var d diff

		switch types[path] {
		case statusFileDeleted:

			deleted := &diffDeleted{}

			deleted.AOID = object.ShortOID(repo.index.entries[path].oid)
			deleted.AMode = string(repo.index.entries[path].permission())
			deleted.APath = filepath.Join("a", path)

			deleted.BOID = object.ShortOID(nullOID)
			deleted.BPath = filepath.Join("b", path)

			d = deleted

		case statusFileModified:

			modified := &diffModified{}

			modified.AOID = object.ShortOID(repo.index.entries[path].oid)
			modified.AMode = string(repo.index.entries[path].permission())
			modified.APath = filepath.Join("a", path)

			f := repo.workspaceFiles[path]

			f.Seek(0, io.SeekStart)
			data, err := io.ReadAll(f)
			if err != nil {
				return diffs, err
			}

			blob, err := object.NewBlob(path, data)
			if err != nil {
				return diffs, err
			}

			modified.BOID = object.ShortOID(blob.OID())
			modified.BMode = string(f.Stats().permission())
			modified.BPath = filepath.Join("b", path)

			d = modified

		}

		diffs = append(diffs, d)

	}

	return diffs, nil
}

func (repo *repository) diffStaged() ([]diff, error) {

	diffs := []diff{}

	files, types := repo.IndexChanges()

	for _, path := range files {

		var d diff

		switch types[path] {
		case statusFileModified:

			modified := &diffModified{}

			f := repo.head[path]

			modified.AOID = object.ShortOID(f.OID())
			modified.AMode = string(f.Permission())
			modified.APath = filepath.Join("b", path)

			modified.BOID = object.ShortOID(repo.index.entries[path].oid)
			modified.BMode = string(repo.index.entries[path].permission())
			modified.BPath = filepath.Join("a", path)

			d = modified

		case statusFileDeleted:

			deleted := &diffDeleted{}

			deleted.AOID = object.ShortOID(repo.head[path].OID())
			deleted.AMode = string(repo.head[path].Permission())
			deleted.APath = filepath.Join("a", path)

			deleted.BOID = object.ShortOID(nullOID)
			deleted.BPath = filepath.Join("b", path)

			d = deleted

		case statusIndexAdded:

			added := &diffAdded{}

			added.AOID = object.ShortOID(nullOID)
			added.APath = filepath.Join("a", path)

			added.BOID = object.ShortOID(repo.index.entries[path].oid)
			added.BMode = string(repo.index.entries[path].permission())
			added.BPath = filepath.Join("b", path)

			d = added

		}

		diffs = append(diffs, d)

	}

	return diffs, nil
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

	repo.headScanner.Walk(func(name string, entry object.Entry) {

		if entry.IsTree() {
			return
		}

		repo.head[name] = entry

		if !repo.index.trackedFile(name) {
			repo.changed[name] = statusFileDeleted + statusNone
			repo.indexChanges[name] = statusFileDeleted
			return
		}

	})

	for _, e := range repo.index.entries {

		indexStatus := statusNone
		if h, ok := repo.head[e.filename]; !ok {
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
