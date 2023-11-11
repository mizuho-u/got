package usecase

import (
	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/io/workspace"
	"github.com/mizuho-u/got/model"
)

func Add(ctx GotContextReaderWriter, paths ...string) error {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	err := db.Index().OpenForUpdate()
	if err != nil {
		return err
	}

	opt := []model.WorkspaceOption{}
	if !db.Index().IsNew() {
		opt = append(opt, model.WithIndex(db.Index()))
	}
	repo, err := model.NewRepository(opt...)
	if err != nil {
		return err
	}

	for _, path := range paths {

		scanner, err := workspace.Scan(ctx.WorkspaceRoot(), path, ctx.GotRoot())
		if err != nil {
			return err
		}

		objects, err := repo.Add(scanner)
		if err != nil {
			return err
		}

		db.Objects().Store(objects...)

	}

	if err := db.Index().Update(repo.Index()); err != nil {
		return err
	}

	return nil
}
