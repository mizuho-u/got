package usecase

import (
	"fmt"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model/object"
)

func ShowHead(ctx GotContext, paths ...string) ExitCode {

	var repo database.Repository = database.NewFS(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer repo.Close()

	head, err := repo.Refs().HEAD()
	if err != nil {
		return 128
	}

	o, err := repo.Objects().Load(head)
	if err != nil {
		return 128
	}

	commit, err := object.ParseCommit(o)
	if err != nil {
		return 128
	}

	repo.Objects().ScanTree(commit.Tree()).Walk(func(name string, entry object.Entry) {
		if entry.IsTree() {
			return
		}

		fmt.Printf("%s %s %s\n", entry.Permission(), entry.OID(), name)
	})

	return 0
}
