package usecase

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
)

func Commit(ctx GotContext, commitMessage string, now time.Time) error {

	filenames, err := listRelativeFilePaths(ctx.WorkspaceRoot(), ctx.GotRoot())
	if err != nil {
		return err
	}

	files := []*model.File{}
	for _, f := range filenames {

		absPath := filepath.Join(ctx.WorkspaceRoot(), f)

		data, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}

		stat, err := os.Stat(absPath)
		if err != nil {
			return err
		}

		files = append(files, &model.File{Name: f, Data: data, Permission: stat.Mode().Perm()})
	}

	refs := database.NewRefs(ctx.GotRoot())
	parent, err := refs.HEAD()
	if err != nil {
		return err
	}

	ws, err := model.NewWorkspace()
	if err != nil {
		return err
	}

	commitId, err := ws.Commit(parent, os.Getenv("GIT_AUTHOR_NAME"), os.Getenv("GIT_AUTHOR_EMAIL"), commitMessage, now, files...)
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

func listRelativeFilePaths(dir, ignore string) ([]string, error) {

	filenames := []string{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(path, ignore) {
			return nil
		}

		path, err = filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		filenames = append(filenames, path)
		return nil

	})

	return filenames, err
}

func msg(parent, commitId, commitMessage string) string {

	prefix := ""
	if parent == "" {
		prefix = "(root-commit) "
	}

	return fmt.Sprintf("[%s%s] %s", prefix, commitId, strings.Split(commitMessage, "\n")[0])

}
