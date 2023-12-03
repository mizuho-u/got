package repository

import (
	"sort"

	"github.com/mizuho-u/got/repository/internal"
)

func (repo *repository) Untracked() []string {

	// 呼び出しのたびにソートするのは無駄かも
	sort.SliceStable(repo.untracked, func(i, j int) bool {
		return repo.untracked[i] < repo.untracked[j]
	})

	return repo.untracked
}

func (repo *repository) Changed() ([]string, map[string]status) {

	files := internal.Keys(repo.changed)

	sort.SliceStable(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	return files, repo.changed
}

func (repo *repository) IndexChanges() ([]string, map[string]status) {

	files := internal.Keys(repo.indexChanges)

	sort.SliceStable(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	return files, repo.indexChanges
}

func (repo *repository) WorkspaceChanges() ([]string, map[string]status) {

	files := internal.Keys(repo.workspaceChanges)

	sort.SliceStable(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	return files, repo.workspaceChanges
}

func (repo *repository) Scan(workspaceScanner WorkspaceScanner, treeScanner TreeScanner) error {

	if err := repo.scan(workspaceScanner); err != nil {
		return err
	}

	repo.detectChanges(treeScanner)

	return nil

}

func (repo *repository) scan(workspaceScanner WorkspaceScanner) error {

	untrackedSet := map[string]struct{}{}

	for {

		p, err := workspaceScanner.Next()
		if err != nil {
			return err
		}

		if p == nil {
			break
		}

		repo.workspace[p.Name()] = p

		if repo.Index().tracked(p.Name()) {
			continue
		}

		entry := p.Name()
		for _, d := range p.Parents() {

			if !repo.Index().tracked(d) {
				entry = d + "/"
				break
			}
		}

		untrackedSet[entry] = struct{}{}
	}

	for k := range untrackedSet {
		repo.untracked = append(repo.untracked, k)
	}

	return nil
}

func (repo *repository) detectChanges(headScanner TreeScanner) {

	headScanner.Walk(func(name string, entry TreeEntry) {

		if entry.IsTree() {
			return
		}

		repo.head[name] = entry

		if !repo.index.trackedFile(name) {
			repo.changed[name] = statusFileDeleted + statusNone
			repo.indexChanges[name] = statusFileDeleted
			return
		}

	})

	for _, e := range repo.index.entries {

		indexStatus := statusNone
		if h, ok := repo.head[e.filename]; !ok {
			indexStatus = statusIndexAdded
			repo.indexChanges[e.filename] = indexStatus
		} else if e.oid != h.OID() || e.permission() != h.Permission() {
			indexStatus = statusFileModified
			repo.indexChanges[e.filename] = indexStatus
		}

		workspaceStatus := statusNone
		if stat, ok := repo.workspace[e.filename]; !ok {
			workspaceStatus = statusFileDeleted
			repo.workspaceChanges[e.filename] = workspaceStatus
		} else if !repo.index.match(stat) {
			workspaceStatus = statusFileModified
			repo.workspaceChanges[e.filename] = workspaceStatus
		}

		status := indexStatus + workspaceStatus
		if status == statusUnchanged {
			continue
		}

		repo.changed[e.filename] = status
	}

}
