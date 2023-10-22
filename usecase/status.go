package usecase

import (
	"fmt"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/model/object"
)

func Status(ctx GotContext) ExitCode {

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
	scanner, err := repo.Scan(ctx.WorkspaceRoot())
	if err != nil {
		ctx.OutError(err)
		return 128
	}
	opt = append(opt, model.WithWorkspaceScanner(scanner))

	head, err := repo.Refs().HEAD()
	if err != nil {
		return 128
	}

	if head != "" {
		o, err := repo.Objects().Load(head)
		if err != nil {
			return 128
		}

		commit, err := object.ParseCommit(o)
		if err != nil {
			return 128
		}
		opt = append(opt, model.WithTreeScanner(repo.Objects().ScanTree(commit.Tree())))
	}

	ws, err := model.NewWorkspace(opt...)
	if err != nil {
		ctx.OutError(err)
		return 128
	}

	if ws.Scan() != nil {
		return 128
	}

	files, types := ws.Changed()
	for _, f := range files {
		ctx.Out(fmt.Sprintf("%s %s\n", types[f], f))
	}

	for _, v := range ws.Untracked() {
		ctx.Out(fmt.Sprintf("?? %s\n", v))
	}

	if err := repo.Index().Update(ws.Index()); err != nil {
		ctx.OutError(err)
		return 128
	}

	return 0
}
