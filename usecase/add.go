package usecase

import (
	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
)

func Add(ctx GotContext, paths ...string) ExitCode {

	var repo database.Repository = database.NewFS(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer repo.Close()

	err := repo.Index().OpenForUpdate()
	if err != nil {
		return 128
	}

	opt := []model.WorkspaceOption{}
	if !repo.Index().IsNew() {
		opt = append(opt, model.WithIndex(repo.Index()))
	}
	ws, err := model.NewWorkspace(opt...)
	if err != nil {
		ctx.OutError(err)
		return 128
	}

	for _, path := range paths {

		scanner, err := repo.Scan(path)
		if err != nil {
			ctx.OutError(err)
			return 128
		}

		blobs, err := ws.Add(scanner)
		if err != nil {
			ctx.OutError(err)
			return 128
		}

		repo.Objects().Store(blobs...)

	}

	if err := repo.Index().Update(ws.Index()); err != nil {
		ctx.OutError(err)
		return 128
	}

	return 0
}
