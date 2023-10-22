package usecase

import (
	"fmt"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/io/database/fs"
	"github.com/mizuho-u/got/model/object"
)

func ShowHead(ctx GotContext, paths ...string) ExitCode {

	var db database.Database = fs.NewFS(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	head, err := db.Refs().HEAD()
	if err != nil {
		return 128
	}

	o, err := db.Objects().Load(head)
	if err != nil {
		return 128
	}

	commit, err := object.ParseCommit(o)
	if err != nil {
		return 128
	}

	db.Objects().ScanTree(commit.Tree()).Walk(func(name string, entry object.Entry) {
		if entry.IsTree() {
			return
		}

		fmt.Printf("%s %s %s\n", entry.Permission(), entry.OID(), name)
	})

	return 0
}
