package usecase

import (
	"github.com/mizuho-u/got/io/database"
)

type ExitCode int

func InitDir(ctx GotContextReaderWriter) error {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())

	if err := db.Init(); err != nil {
		return err
	}

	ctx.Out("Initialized empty Jit repository in "+ctx.WorkspaceRoot(), none)

	return nil
}
