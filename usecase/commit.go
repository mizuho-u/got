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

func Commit(wd, message string, now time.Time) (commitId string, err error) {

	gotpath := filepath.Join(wd, ".git")

	filenames, err := listRelativeFilePaths(wd, gotpath)
	if err != nil {
		return "", err
	}

	files := []*model.File{}
	for _, f := range filenames {

		absPath := filepath.Join(wd, f)

		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", err
		}

		stat, err := os.Stat(absPath)
		if err != nil {
			return "", err
		}

		files = append(files, &model.File{Name: f, Data: data, Permission: stat.Mode().Perm()})
	}

	refs := database.NewRefs(gotpath)
	parent, err := refs.HEAD()
	if err != nil {
		return "", err
	}

	ws, err := model.NewWorkspace()
	if err != nil {
		return "", err
	}

	commitId, err = ws.Commit(parent, os.Getenv("GIT_AUTHOR_NAME"), os.Getenv("GIT_AUTHOR_EMAIL"), message, now, files...)
	if err != nil {
		return "", err
	}

	objects := database.NewObjects(gotpath)
	objects.StoreAll(ws.Objects()...)

	if err := refs.UpdateHEAD(commitId); err != nil {
		return "", err
	}

	prefix := ""
	if parent == "" {
		prefix = "(root-commit) "
	}
	fmt.Printf("%s%s %s", prefix, commitId, message)

	return commitId, nil

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
