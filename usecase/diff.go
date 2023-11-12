package usecase

import (
	"fmt"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/io/workspace"
	"github.com/mizuho-u/got/repository"
)

func Diff(ctx GotContextReaderWriter, staged bool) error {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	err := db.Index().OpenForUpdate()
	if err != nil {
		return err
	}

	opt := []repository.WorkspaceOption{}
	if !db.Index().IsNew() {
		opt = append(opt, repository.WithIndex(db.Index()))
	}

	opt = append(opt, repository.WithObjectLoader(db.Objects()))

	repo, err := repository.NewRepository(opt...)
	if err != nil {
		return err
	}

	scanner, err := workspace.Scan(ctx.WorkspaceRoot(), ctx.WorkspaceRoot(), ctx.GotRoot())
	if err != nil {
		return err
	}
	head, err := db.Refs().Head()
	if err != nil {
		return err
	}

	if repo.Scan(scanner, db.Objects().ScanTree(head.Tree())) != nil {
		return err
	}

	diffs, err := repo.Diff(staged)
	if err != nil {
		return err
	}

	for _, diff := range diffs {

		ctx.Out(diff.PathLine(), bold)
		ctx.Out(diff.ModeLine(), bold)
		ctx.Out(diff.IndexLine(), bold)
		ctx.Out(diff.FileLine(), bold)

		for _, hunk := range diff.Hunks() {

			ctx.Out(fmt.Sprintln(hunk.Header()), cyan)

			for _, edit := range hunk.Edits() {

				color := none
				switch edit.Diff() {
				case repository.Deletion:
					color = red
				case repository.Insertion:
					color = green
				default:
				}

				ctx.Out(fmt.Sprintln(edit), color)
			}

		}

	}

	if err := db.Index().Update(repo.Index()); err != nil {
		return err
	}

	return nil
}
