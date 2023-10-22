package usecase

import (
	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/io/workspace"
	"github.com/mizuho-u/got/model"
)

func Add(ctx GotContext, paths ...string) ExitCode {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	err := db.Index().OpenForUpdate()
	if err != nil {
		return 128
	}

	opt := []model.WorkspaceOption{}
	if !db.Index().IsNew() {
		opt = append(opt, model.WithIndex(db.Index()))
	}
	ws, err := model.NewWorkspace(opt...)
	if err != nil {
		ctx.OutError(err)
		return 128
	}

	for _, path := range paths {

		scanner, err := workspace.Scan(ctx.WorkspaceRoot(), path, ctx.GotRoot())
		if err != nil {
			ctx.OutError(err)
			return 128
		}

		blobs, err := ws.Add(scanner)
		if err != nil {
			ctx.OutError(err)
			return 128
		}

		db.Objects().Store(blobs...)

	}

	if err := db.Index().Update(ws.Index()); err != nil {
		ctx.OutError(err)
		return 128
	}

	return 0
}
