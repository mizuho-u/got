package usecase

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/usecase/internal"
)

func Status(ctx GotContext) ExitCode {

	var repo database.Repository = database.NewFS(ctx.GotRoot())
	defer repo.Close()

	err := repo.Index().OpenForRead()
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

	paths, err := internal.ListRelativeFilePaths(ctx.WorkspaceRoot(), ctx.GotRoot())
	if err != nil {
		ctx.OutError(err)
		return 128
	}

	untracked := []string{}

	for _, p := range paths {

		if ws.Index().Tracked(p) {
			continue
		}

		if p == filepath.Base(p) {
			untracked = append(untracked, p)
			continue
		}

		if ws.Index().Tracked(filepath.Dir(p)) {
			untracked = append(untracked, p)
		} else {
			untracked = append(untracked, filepath.Dir(p)+"/")
		}

	}

	sort.SliceStable(untracked, func(i, j int) bool {
		return untracked[i] < untracked[j]
	})

	for _, v := range untracked {
		ctx.Out(fmt.Sprintf("?? %s\n", v))
	}

	return 0
}
