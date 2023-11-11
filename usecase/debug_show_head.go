package usecase

import (
	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/model"
)

func ShowHead(ctx GotContext, paths ...string) error {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	head, err := db.Refs().Head()
	if err != nil {
		return err
	}

	tree := ""
	if head != nil {
		tree = head.Tree()
	}

	db.Objects().ScanTree(tree).Walk(func(name string, entry model.TreeEntry) {
		if entry.IsTree() {
			return
		}

		// fmt.Printf("%s %s %s\n", entry.Permission(), entry.OID(), name)
	})

	return nil
}
