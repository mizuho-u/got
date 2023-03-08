package usecase

import (
	"io/fs"
	"log"
	"path/filepath"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/usecase/internal"
)

func Add(wd string, paths ...string) error {

	gotpath := filepath.Join(wd, ".git")

	filepaths, err := internal.ListFilepathsRecursively(paths, filepath.Join(gotpath))
	if err != nil {
		return err
	}

	files := []*model.File{}
	for _, path := range filepaths {

		stat, err := internal.FileStat(path)
		if err != nil {
			return err
		}

		data, err := internal.ReadFile(path)
		if err != nil {
			return err
		}

		relpath, err := filepath.Rel(wd, path)
		if err != nil {
			log.Println(wd, path)
			return err
		}

		files = append(files, &model.File{
			Name:       relpath,
			Data:       data,
			Permission: internal.FileModePerm(fs.FileMode(stat.Mode)),
			Stat:       model.NewFileStat(stat),
		})

	}

	index, err := database.OpenIndex(gotpath)
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
		return err
	}
	ws.Add(files...)

	objects := database.NewObjects(gotpath)
	objects.StoreAll(ws.Objects()...)

	index.Update(ws.Index())

	return nil
}
