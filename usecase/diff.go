package usecase

import (
	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/io/workspace"
	"github.com/mizuho-u/got/model"
)

func Diff(ctx GotContext, staged bool) ExitCode {

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

	head, err := db.Refs().Head()
	if err != nil {
		return 128
	}
	opt = append(opt, model.WithTreeScanner(db.Objects().ScanTree(head.Tree())))

	repo, err := model.NewRepository(opt...)
	if err != nil {
		ctx.OutError(err)
		return 128
	}

	if repo.Scan() != nil {
		return 128
	}

	diffs, err := repo.Diff(staged)
	if err != nil {
		return 128
	}

	for _, diff := range diffs {

		ctx.Out(diff.PathLine(), none)
		ctx.Out(diff.ModeLine(), none)
		ctx.Out(diff.IndexLine(), none)
		ctx.Out(diff.FileLine(), none)

	}

	if err := db.Index().Update(repo.Index()); err != nil {
		ctx.OutError(err)
		return 128
	}

	return 0
}
