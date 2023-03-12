package usecase

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
)

func Commit(ctx GotContext, commitMessage string, now time.Time) error {

	index, err := database.OpenIndexForRead(ctx.GotRoot())
	if err != nil {
		return err
	}

	ws, err := model.NewWorkspace(model.WithIndex(index))
	if err != nil {
		return err
	}

	refs := database.NewRefs(ctx.GotRoot())
	parent, err := refs.HEAD()
	if err != nil {
		return err
	}

	commitId, err := ws.Commit(parent, os.Getenv("GIT_AUTHOR_NAME"), os.Getenv("GIT_AUTHOR_EMAIL"), commitMessage, now)
	if err != nil {
		return err
	}

	objects := database.NewObjects(ctx.GotRoot())
	objects.StoreAll(ws.Objects()...)

	if err := refs.UpdateHEAD(commitId); err != nil {
		return err
	}

	if err := ctx.Out(msg(parent, commitId, commitMessage)); err != nil {
		return err
	}

	return nil
}

func msg(parent, commitId, commitMessage string) string {

	prefix := ""
	if parent == "" {
		prefix = "(root-commit) "
	}

	return fmt.Sprintf("[%s%s] %s", prefix, commitId, strings.Split(commitMessage, "\n")[0])

}
