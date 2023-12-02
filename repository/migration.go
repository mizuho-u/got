package repository

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mizuho-u/got/internal"
	"github.com/mizuho-u/got/internal/statement"
	"github.com/mizuho-u/got/repository/object"
)

type migrateAction int

const (
	migrateCreate migrateAction = iota
	migrateUpdate
	migrateDelete
)

type migrateConflict int

const (
	conflictStaleFile migrateConflict = iota
	conflictStaleDirectory
	conflictUntrackedOverwritten
	conflictUntrackedRemoved
)

type migrationChanges map[migrateAction][]internal.Tuple[string, object.TreeEntry]

type migration struct {
	diff      map[string]pair
	changes   migrationChanges
	mkdirs    internal.Set[string]
	rmdirs    internal.Set[string]
	ws        Workspace
	db        Database
	index     IndexWriteReader
	inspector Inspector
	conflicts map[migrateConflict]internal.Set[string]
}

func NewMigration(diff map[string]pair, ws Workspace, db Database, index IndexWriteReader, inspector Inspector) *migration {

	m := &migration{
		diff:      diff,
		changes:   make(migrationChanges),
		mkdirs:    internal.NewSet[string](),
		rmdirs:    internal.NewSet[string](),
		ws:        ws,
		db:        db,
		index:     index,
		inspector: inspector,
		conflicts: make(map[migrateConflict]internal.Set[string])}

	m.conflicts[conflictStaleFile] = internal.NewSet[string]()
	m.conflicts[conflictStaleDirectory] = internal.NewSet[string]()
	m.conflicts[conflictUntrackedOverwritten] = internal.NewSet[string]()
	m.conflicts[conflictUntrackedRemoved] = internal.NewSet[string]()

	return m
}

func (m *migration) ApplyChanges() error {

	if err := m.planChanges(); err != nil {
		return err
	}

	if err := m.updateWorkspace(); err != nil {
		return err
	}

	if err := m.updateIndex(); err != nil {
		return err
	}

	return nil

}

func (m *migration) planChanges() error {

	for path, pair := range m.diff {

		if err := m.checkForConfilct(path, pair.Item1(), pair.Item2()); err != nil {
			return err
		}

		dirs := internal.NewSetFromArray(internal.ParentDirs(path))

		if pair.Item1() == nil {

			m.mkdirs.Merge(dirs)
			m.changes[migrateCreate] = append(m.changes[migrateCreate], internal.NewTuple(path, pair.Item2()))

		} else if pair.Item2() == nil {

			m.rmdirs.Merge(dirs)
			m.changes[migrateDelete] = append(m.changes[migrateDelete], internal.NewTuple(path, pair.Item2()))

		} else {

			m.changes[migrateUpdate] = append(m.changes[migrateUpdate], internal.NewTuple(path, pair.Item2()))

		}

	}

	return nil
}

func (m *migration) updateWorkspace() error {

	if err := m.removeWorkspaceFiles(); err != nil {
		return err
	}

	if err := m.removeEmptyWorkspaceDirs(); err != nil {
		return err
	}

	if err := m.createWorkspaceDirs(); err != nil {
		return err
	}

	if err := m.updateWorkspaceFiles(); err != nil {
		return err
	}

	if err := m.createWorkspaceFiles(); err != nil {
		return err
	}

	return nil

}

func (m *migration) removeWorkspaceFiles() error {

	for _, f := range m.changes[migrateDelete] {
		if err := m.ws.RemoveFile(f.Item1()); err != nil {
			return err
		}
	}

	return nil

}

func (m *migration) removeEmptyWorkspaceDirs() error {

	dirs := m.rmdirs.Iter()
	sort.SliceStable(dirs, func(i, j int) bool {
		return dirs[i] > dirs[j]
	})

	for _, f := range dirs {
		m.ws.RemoveDirectory(f)
	}

	return nil

}

func (m *migration) createWorkspaceDirs() error {

	dirs := m.mkdirs.Iter()
	sort.SliceStable(dirs, func(i, j int) bool {
		return dirs[i] < dirs[j]
	})

	for _, d := range dirs {

		if stat, err := m.ws.Stat(d); err == nil && !stat.IsDir() {
			if err := m.ws.RemoveFile(d); err != nil {
				return err
			}
		}

		if err := m.ws.CreateDir(d); err != nil {
			return err
		}
	}

	return nil

}

func (m *migration) updateWorkspaceFiles() error {

	for _, f := range m.changes[migrateUpdate] {

		o, err := m.db.Load(f.Item2().OID())
		if err != nil {
			return err
		}

		modify, err := m.ws.Open(f.Item1())
		if err != nil {
			return err
		}

		if _, err := modify.Write(o.Data()); err != nil {
			return err
		}

		if err := modify.Chmod(f.Item2().Permission()); err != nil {
			return err
		}

		if err := modify.Close(); err != nil {
			return err
		}

	}

	return nil

}

func (m *migration) createWorkspaceFiles() error {

	for _, f := range m.changes[migrateCreate] {

		new, err := m.ws.CreateFile(f.Item1())
		if err != nil {
			return err
		}

		o, err := m.db.Load(f.Item2().OID())
		if err != nil {
			return err
		}

		if _, err := new.Write(o.Data()); err != nil {
			return err
		}

		if err := new.Chmod(f.Item2().Permission()); err != nil {
			return err
		}

		if err := new.Close(); err != nil {
			return err
		}

	}

	return nil

}

func (m *migration) updateIndex() error {

	for _, change := range m.changes[migrateDelete] {
		m.index.Delete(change.Item1())
	}

	for _, change := range m.changes[migrateCreate] {
		stat, err := m.ws.Stat(change.Item1())
		if err != nil {
			return err
		}
		m.index.Add(NewIndexEntry(change.Item1(), change.Item2().OID(), stat.Stats()))
	}

	for _, change := range m.changes[migrateUpdate] {
		stat, err := m.ws.Stat(change.Item1())
		if err != nil {
			return err
		}
		m.index.Add(NewIndexEntry(change.Item1(), change.Item2().OID(), stat.Stats()))
	}

	return nil

}

func (m *migration) checkForConfilct(path string, old, new object.TreeEntry) error {

	ie, ok := m.index.Get(path)

	// indexにあるけどtreeにないファイルはstalefile
	if ok && m.indexDiffersFromTrees(ie, old, new) {
		m.conflicts[conflictStaleFile].Set(path)
		return nil
	}

	stat, err := m.ws.Stat(path)
	conflict := m.getConflictType(ie, stat, new)

	// wsにファイルが存在しない場合は
	if err != nil {

		parent := m.untrackedParent(path)
		if parent != "" {
			m.conflicts[conflict].Set(statement.Ternary(ok, path, parent))
		}

	} else if !stat.IsDir() {
		f, err := m.ws.Open(path)
		if err != nil {
			return err
		}

		changed := m.inspector.CompareIndexToWorkspace(ie, f)
		if changed != statusNone {
			m.conflicts[conflict].Set(path)
		}

	} else {

		trackable, err := m.inspector.TrackableFile(path, stat)
		if err != nil {
			return err
		}

		if trackable {
			m.conflicts[conflict].Set(path)
		}

	}

	return nil

}

func (m *migration) getConflictType(ie *IndexEntry, stat WorkspaceFileStat, te object.TreeEntry) migrateConflict {

	if ie != nil {
		return conflictStaleFile
	}

	if stat != nil && stat.IsDir() {
		return conflictStaleDirectory
	}

	if te != nil {
		return conflictUntrackedOverwritten
	}

	return conflictUntrackedRemoved

}

func (m *migration) untrackedParent(path string) string {

	for _, parent := range internal.ParentDirs(path) {

		stat, err := m.ws.Stat(parent)
		if err != nil || stat.IsDir() {
			continue
		}

		if trackable, err := m.inspector.TrackableFile(parent, stat); trackable && err == nil {
			return parent
		}

	}

	return ""
}

func (m *migration) indexDiffersFromTrees(entry *IndexEntry, old, new object.TreeEntry) bool {

	if m.inspector.CompareTreeToIndex(old, entry) != statusNone && m.inspector.CompareTreeToIndex(new, entry) != statusNone {
		return true
	}

	return false

}

var conflictMessages map[migrateConflict]internal.Tuple[string, string] = map[migrateConflict]internal.Tuple[string, string]{
	conflictStaleFile: internal.NewTuple(
		"Your local changes to the following files would be overwritten by checkout:",
		"Please commit your changes or stash them before you switch branches."),
	conflictStaleDirectory: internal.NewTuple(
		"Updating the following directoris would lose untracked files in them:", "\n"),
	conflictUntrackedOverwritten: internal.NewTuple(
		"The following untracked working tree files would be overwritten by checkout:",
		"Please move or remove them before you switch branches"),
	conflictUntrackedRemoved: internal.NewTuple(
		"The following untracked working tree files would be removed by checkout:",
		"Please move or remove them before you switch branches"),
}

func (m *migration) Conflicts() []error {

	errors := []error{}
	for t, conflicts := range m.conflicts {

		if conflicts.Length() == 0 {
			continue
		}

		msg := conflictMessages[t]

		errors = append(errors, fmt.Errorf("%s\n\t%s\n%s", msg.Item1(), strings.Join(conflicts.Iter(), "\n"), msg.Item2()))

	}

	return errors

}
