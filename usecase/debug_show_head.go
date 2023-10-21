package usecase

import (
	"fmt"
	"path/filepath"

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

	showTree(repo, commit.Tree(), "")

	return 0
}

func showTree(repo database.Repository, oid, path string) {

	o, err := repo.Objects().Load(oid)
	if err != nil {
		return
	}

	tree, err := object.ParseTree(o)
	if err != nil {
		return
	}

	for _, entry := range tree.Children() {

		if entry.IsTree() {
			showTree(repo, entry.OID(), filepath.Join(path, entry.Basename()))
		} else {
			fmt.Printf("%s %s %s\n", entry.Permission(), entry.OID(), filepath.Join(path, entry.Basename()))
		}

	}

}
