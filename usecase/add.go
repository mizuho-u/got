package usecase

import (
	"io/fs"
	"log"
	"path/filepath"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/usecase/internal"
)

func Add(ctx GotContext, paths ...string) error {

	filepaths, err := internal.ListFilepathsRecursively(paths, ctx.GotRoot())
	if err != nil {
		return wrap(err)
	}

	files := []*model.File{}
	for _, path := range filepaths {

		stat, err := internal.FileStat(path)
		if err != nil {
			return wrap(err)
		}

		data, err := internal.ReadFile(path)
		if err != nil {
			return wrap(err)
		}

		relpath, err := filepath.Rel(ctx.WorkspaceRoot(), path)
		if err != nil {
			log.Println(ctx.WorkspaceRoot(), path)
			return wrap(err)
		}

		files = append(files, &model.File{
			Name:       relpath,
			Data:       data,
			Permission: internal.FileModePerm(fs.FileMode(stat.Mode)),
			Stat:       model.NewFileStat(stat),
		})

	}

	index, err := database.OpenIndexForUpdate(ctx.GotRoot())
	if err != nil {
		return err
	}
	defer index.Close()

	opt := []model.WorkspaceOption{}
	if !index.IsNew() {
		opt = append(opt, model.WithIndex(index))
	}

	ws, err := model.NewWorkspace(opt...)
	if err != nil {
		return wrap(err)
	}
	ws.Add(files...)

	objects := database.NewObjects(ctx.GotRoot())
	objects.StoreAll(ws.Objects()...)

	index.Update(ws.Index())

	return nil
}
