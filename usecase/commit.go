package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
)

func Commit(ctx GotContext, commitMessage string, now time.Time) ExitCode {

	var repo database.Repository = database.NewFS(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer repo.Close()

	err := repo.Index().OpenForRead()
	if err != nil {
		return 128
	}

	ws, err := model.NewWorkspace(model.WithIndex(repo.Index()))
	if err != nil {
		return 128
	}

	parent, err := repo.Refs().HEAD()
	if err != nil {
		return 128
	}

	commitId, err := ws.Commit(parent, ctx.Username(), ctx.Email(), commitMessage, now)
	if err != nil {
		return 128
	}

	if err := repo.Refs().UpdateHEAD(commitId); err != nil {
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
