package usecase

import (
	"errors"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/io/workspace"
	"github.com/mizuho-u/got/repository"
	"github.com/mizuho-u/got/types"
)

func Checkout(ctx GotContextReaderWriter, revision types.Revision) error {

	ws := workspace.New(ctx.WorkspaceRoot())

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	err := db.Index().OpenForUpdate()
	if err != nil {
		return err
	}

	oid, err := revision.Resolve(&resolver{refs: db.Refs(), objects: db.Objects()})
	if err != nil {
		return err
	}

	head, err := db.Refs().Head()
	if err != nil {
		return err
	}

	diff := repository.NewTreeDiff(db.Objects())
	if err := diff.Diff(types.ObjectID(head.OID()), oid); err != nil {
		return err
	}

	index, err := repository.NewIndex(repository.IndexSource(db.Index()))
	if err != nil {
		return err
	}

	m := repository.NewMigration(diff.Changes(), ws, db.Objects(), index, repository.NewInspector(index, ws))
	if err := m.ApplyChanges(); err != nil {
		return err
	}

	if conflicts := m.Conflicts(); len(conflicts) != 0 {
		return errors.Join(conflicts...)
	}

	if err := db.Refs().UpdateHeadCommit(oid.String()); err != nil {
		return err
	}

	if err := db.Index().Update(index); err != nil {
		return err
	}

	return nil

}
