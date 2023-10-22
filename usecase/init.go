package usecase

import (
	"github.com/mizuho-u/got/io/database"
)

type ExitCode int

func InitDir(ctx GotContext) ExitCode {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())

	if err := db.Init(); err != nil {
		return 128
	}

	ctx.Out("Initialized empty Jit repository in " + ctx.WorkspaceRoot())

	return 0

}
