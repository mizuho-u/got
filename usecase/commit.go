package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/model"
)

func Commit(ctx GotContext, commitMessage string, now time.Time) ExitCode {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	err := db.Index().OpenForRead()
	if err != nil {
		return 128
	}

	ws, err := model.NewWorkspace(model.WithIndex(db.Index()))
	if err != nil {
		return 128
	}

	parent, err := db.Refs().HEAD()
	if err != nil {
		return 128
	}

	commitId, err := ws.Commit(parent, ctx.Username(), ctx.Email(), commitMessage, now)
	if err != nil {
		return 128
	}

	if err := db.Objects().Store(ws.Objects()...); err != nil {
		return 128
	}

	if err := db.Refs().UpdateHEAD(commitId); err != nil {
		return 128
	}

	if err := ctx.Out(msg(parent, commitId, commitMessage)); err != nil {
		return 128
	}

	return 0
}

func msg(parent, commitId, commitMessage string) string {

	prefix := ""
	if parent == "" {
		prefix = "(root-commit) "
	}

	return fmt.Sprintf("[%s%s] %s", prefix, commitId, strings.Split(commitMessage, "\n")[0])

}
