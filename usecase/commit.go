package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/model"
)

func Commit(ctx GotContext, commitMessage string, now time.Time) error {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	err := db.Index().OpenForRead()
	if err != nil {
		return err
	}

	repo, err := model.NewRepository(model.WithIndex(db.Index()))
	if err != nil {
		return err
	}

	head, err := db.Refs().Head()
	if err != nil {
		return err
	}

	parent := head.OID()

	commitId, objects, err := repo.Commit(parent, ctx.Username(), ctx.Email(), commitMessage, now)
	if err != nil {
		return err
	}

	if err := db.Objects().Store(objects...); err != nil {
		return err
	}

	if err := db.Refs().UpdateHEAD(commitId); err != nil {
		return err
	}

	if err := ctx.Out(msg(parent, commitId, commitMessage), none); err != nil {
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
