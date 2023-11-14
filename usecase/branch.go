package usecase

import (
	"errors"
	"fmt"

	"github.com/mizuho-u/got/io/database"
	"github.com/mizuho-u/got/repository/object"
	"github.com/mizuho-u/got/types"
)

func Branch(ctx GotContextReaderWriter, branchName types.BranchName, startPoint types.Revision) error {

	var db database.Database = database.NewFSDB(ctx.WorkspaceRoot(), ctx.GotRoot())
	defer db.Close()

	sp, err := startPoint.Resolve(&resolver{refs: db.Refs(), objects: db.Objects()})
	if err != nil {
		return err
	}

	return db.Refs().CreateBranch(branchName, sp.String())
}

type resolver struct {
	refs    database.Refs
	objects database.Objects
}

func (r *resolver) Ref(name string) (types.ObjectID, error) {

	commit, err := r.refs.Ref(name)
	if err == nil {
		return types.NewObjectID(commit.OID())
	}

	objects, err := r.objects.LoadPrefix(name)

	if err != nil || len(objects) == 0 {

		return "", fmt.Errorf("not a valid object name: %s", name)

	} else if len(objects) == 1 {

		c, err := object.ParseCommit(objects[0])
		if err == nil {
			return types.NewObjectID(c.OID())
		} else {
			return "", fmt.Errorf("object %s is a %s, not a commit", objects[0].OID(), objects[0].Class())
		}

	} else {
		hints := fmt.Sprintf("short SHA1 %s is ambiguous\n", name)

		for _, o := range objects {

			if c, err := object.ParseCommit(o); err == nil {
				hints += fmt.Sprintf("hint: %s %s %s\n", object.ShortOID(o.OID()), o.Class(), c.TitleLine())
			} else {
				hints += fmt.Sprintf("hint: %s %s\n", object.ShortOID(o.OID()), o.Class())
			}
		}

		return "", errors.New(hints)

	}

}

func (r *resolver) Parent(oid types.ObjectID) (types.ObjectID, error) {

	commit, err := r.objects.LoadCommit(oid.String())
	if err != nil {
		return "", err
	}

	if commit.Parent() == "" {
		return "", fmt.Errorf("no parents found: %s", oid)
	}

	return types.NewObjectID(commit.Parent())
}
