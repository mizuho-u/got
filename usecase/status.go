package usecase

import (
	"fmt"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
)

func Status(ctx GotContext) ExitCode {

	var repo database.Repository = database.NewFS(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer repo.Close()

	err := repo.Index().OpenForRead()
	if err != nil {
		return 128
	}

	opt := []model.WorkspaceOption{}
	if !repo.Index().IsNew() {
		opt = append(opt, model.WithIndex(repo.Index()))
	}
	opt = append(opt, model.WithWorkspaceScanner(repo.Scan()))

	ws, err := model.NewWorkspace(opt...)
	if err != nil {
		ctx.OutError(err)
		return 128
	}

	if ws.Scan() != nil {
		return 128
	}

	for _, v := range ws.Untracked() {
		ctx.Out(fmt.Sprintf("?? %s\n", v))
	}

	return 0
}
