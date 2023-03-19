package usecase

import (
	"io/fs"
	"path/filepath"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/usecase/internal"
)

func Add(ctx GotContext, paths ...string) error {

	index, err := database.OpenIndexForUpdate(ctx.GotRoot())
	if err != nil {
		return err
	}
	defer index.Close()

	objects := database.NewObjects(ctx.GotRoot())

	opt := []model.WorkspaceOption{}
	if !index.IsNew() {
		opt = append(opt, model.WithIndex(index))
	}
	ws, err := model.NewWorkspace(opt...)
	if err != nil {
		return err
	}

	for _, path := range paths {

		files, err := readFilesToAdd(ctx, path)
		if err != nil {
			return err
		}

		for _, f := range files {

			blob, err := ws.Add(f)
			if err != nil {
				return err
			}

			objects.Store(blob)
		}

	}

	if err := index.Update(ws.Index()); err != nil {
		return err
	}

	return nil
}

func readFilesToAdd(ctx GotContext, path string) ([]*model.File, error) {

	paths := []string{}
	paths = append(paths, path)

	filepaths, err := internal.ListFilepathsRecursively(paths, ctx.GotRoot())
	if err != nil {
		return nil, err
	}

	files := []*model.File{}
	for _, p := range filepaths {

		stat, err := internal.FileStat(p)
		if err != nil {
			return nil, err
		}

		data, err := internal.ReadFile(p)
		if err != nil {
			return nil, err
		}

		relpath, err := filepath.Rel(ctx.WorkspaceRoot(), p)
		if err != nil {
			return nil, err
		}

		files = append(files, &model.File{
			Name:       relpath,
			Data:       data,
			Permission: internal.FileModePerm(fs.FileMode(stat.Mode)),
			Stat:       model.NewFileStat(stat),
		})

	}

	return files, nil

}
