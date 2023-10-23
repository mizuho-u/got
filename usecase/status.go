package usecase

import (
	"fmt"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/io/workspace"
	"github.com/mizuho-u/got/model"
)

func Status(ctx GotContext, porcelain bool) ExitCode {

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

	if porcelain {
		files, types := repo.Changed()
		for _, f := range files {
			ctx.Out(fmt.Sprintf("%s %s\n", types[f].ShortFormat(), f))
		}

		for _, v := range repo.Untracked() {
			ctx.Out(fmt.Sprintf("?? %s\n", v))
		}
	} else {

		indexChanges := false
		if files, types := repo.IndexChanges(); len(files) != 0 {
			ctx.Out("Changes to be commited:\n\n")
			for _, f := range files {
				ctx.Out(fmt.Sprintf("\t%8s: %s\n", types[f].LongFormat(), f))
			}
			ctx.Out("\n")

			indexChanges = true
		}

		workspaceChanges := false
		if files, types := repo.WorkspaceChanges(); len(files) != 0 {
			ctx.Out("Changes not staged for commit:\n\n")
			for _, f := range files {
				ctx.Out(fmt.Sprintf("\t%8s: %s\n", types[f].LongFormat(), f))
			}
			ctx.Out("\n")

			workspaceChanges = true
		}

		untrackedFiles := false
		if files := repo.Untracked(); len(files) != 0 {
			ctx.Out("Untracked files:\n\n")
			for _, f := range files {
				ctx.Out(fmt.Sprintf("\t%-8s\n", f))
			}
			ctx.Out("\n")

			untrackedFiles = true
		}

		if !indexChanges {

			if workspaceChanges {
				ctx.Out("no changes added to commit")
			} else if untrackedFiles {
				ctx.Out("nothing added to commit but untracked files present")
			} else {
				ctx.Out("nothing to commit, working tree clean")
			}

		}

	}

	if err := db.Index().Update(repo.Index()); err != nil {
		ctx.OutError(err)
		return 128
	}

	return 0
}
