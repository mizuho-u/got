package usecase

import (
	"fmt"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/io/workspace"
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/model/object"
)

func Status(ctx GotContext) ExitCode {

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
	scanner, err := workspace.Scan(ctx.WorkspaceRoot(), ctx.WorkspaceRoot(), ctx.GotRoot())
	if err != nil {
		ctx.OutError(err)
		return 128
	}
	opt = append(opt, model.WithWorkspaceScanner(scanner))

	head, err := db.Refs().HEAD()
	if err != nil {
		return 128
	}

	if head != "" {
		o, err := db.Objects().Load(head)
		if err != nil {
			return 128
		}

		commit, err := object.ParseCommit(o)
		if err != nil {
			return 128
		}
		opt = append(opt, model.WithTreeScanner(db.Objects().ScanTree(commit.Tree())))
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

	if err := db.Index().Update(ws.Index()); err != nil {
		ctx.OutError(err)
		return 128
	}

	return 0
}
