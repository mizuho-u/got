package usecase

import (
	"io/fs"
	"path/filepath"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/usecase/internal"
)

func Add(ctx GotContext, paths ...string) ExitCode {

	var repo database.Repository = database.NewFS(ctx.GotRoot())
	defer repo.Close()

	err := repo.Index().OpenForUpdate()
	if err != nil {
		return 128
	}

	opt := []model.WorkspaceOption{}
	if !repo.Index().IsNew() {
		opt = append(opt, model.WithIndex(repo.Index()))
	}
	ws, err := model.NewWorkspace(opt...)
	if err != nil {
		ctx.OutError(err)
		return 128
	}

	for _, path := range paths {

		files, err := readFilesToAdd(ctx, path)
		if err != nil {
			ctx.OutError(err)
			return 128
		}

		for _, f := range files {

			blob, err := ws.Add(f)
			if err != nil {
				ctx.OutError(err)
				return 128
			}

			repo.Objects().Store(blob)
		}

	}

	if err := repo.Index().Update(ws.Index()); err != nil {
		ctx.OutError(err)
		return 128
	}

	return 0
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
