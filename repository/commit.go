package repository

import (
	"time"

	"github.com/mizuho-u/got/repository/object"
)

func (repo *repository) Commit(parent, author, email, message string, now time.Time) (commitId string, objects []object.Object, err error) {

	entries := []object.TreeEntry{}

	for _, entry := range repo.index.entries {
		entries = append(entries, object.NewTreeEntry(entry.filename, entry.permission(), entry.oid))
	}

	root, err := object.BuildTree(entries)
	if err != nil {
		return commitId, objects, err
	}

	root.Walk(func(tree object.Object) error {

		objects = append(objects, tree)
		return nil

	})

	a := object.NewAuthor(author, email, now)
	commit, err := object.NewCommit(parent, root.OID(), a, message)
	if err != nil {
		return commitId, objects, err
	}
	objects = append(objects, commit)

	commitId = commit.OID()

	return commitId, objects, err

}
