package repository

import (
	"sort"

	"github.com/mizuho-u/got/internal"
	"github.com/mizuho-u/got/repository/object"
)

type migrateAction int

const (
	migrateCreate migrateAction = iota
	migrateUpdate
	migrateDelete
)

type migrationChanges map[migrateAction][]internal.Tuple[string, object.TreeEntry]

type migration struct {
	diff    map[string]pair
	changes migrationChanges
	mkdirs  internal.Set[string]
	rmdirs  internal.Set[string]
	ws      Workspace
	db      Database
	index   IndexWriter
}

func NewMigration(diff map[string]pair, ws Workspace, db Database, index IndexWriter) *migration {
	return &migration{diff: diff, changes: make(migrationChanges), mkdirs: internal.NewSet[string](), rmdirs: internal.NewSet[string](), ws: ws, db: db, index: index}
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
