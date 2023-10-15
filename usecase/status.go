package usecase

import (
	"fmt"
	"sort"

	"github.com/mizuho-u/got/database"
	"github.com/mizuho-u/got/model"
	"github.com/mizuho-u/got/usecase/internal"

	gotinternal "github.com/mizuho-u/got/internal"
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

	untrackedSet := map[string]struct{}{}
	for _, p := range paths {

		if ws.Index().Tracked(p) {
			continue
		}

		entry := p
		for _, d := range gotinternal.ParentDirs(p) {

			if !ws.Index().Tracked(d) {
				entry = d + "/"
				break
			}
		}

		untrackedSet[entry] = struct{}{}
	}

	untracked := make([]string, 0, len(untrackedSet))
	for k := range untrackedSet {
		untracked = append(untracked, k)
	}

	sort.SliceStable(untracked, func(i, j int) bool {
		return untracked[i] < untracked[j]
	})

	for _, v := range untracked {
		ctx.Out(fmt.Sprintf("?? %s\n", v))
	}

	return 0
}
