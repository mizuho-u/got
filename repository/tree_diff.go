package repository

import (
	"errors"
	"path/filepath"

	"github.com/mizuho-u/got/internal"
	"github.com/mizuho-u/got/repository/object"
	"github.com/mizuho-u/got/types"
)

type objectLoader interface {
	Load(oid string) (object.Object, error)
}

type pair internal.Tuple[object.TreeEntry, object.TreeEntry]

func newPair(a, b object.TreeEntry) pair {
	return internal.NewTuple(a, b)
}

type treeDiff struct {
	ol      objectLoader
	changes map[string]pair
}

func NewTreeDiff(ol objectLoader) *treeDiff {
	return &treeDiff{ol, map[string]pair{}}
}

func (td *treeDiff) Diff(a, b types.ObjectID) error {
	return td.compare(a, b, "")
}

func (td *treeDiff) compare(a, b types.ObjectID, prefix string) error {

	if a == b {
		return nil
	}

	atree := map[string]object.TreeEntry{}
	if tree, err := td.loadTree(a); err == nil {
		atree = tree.ChildrenMap()
	}

	btree := map[string]object.TreeEntry{}
	if tree, err := td.loadTree(b); err == nil {
		btree = tree.ChildrenMap()
	}

	if err := td.detectDeletion(atree, btree, prefix); err != nil {
		return err
	}

	if err := td.detectAdditions(atree, btree, prefix); err != nil {
		return err
	}

	return nil
}

func (td *treeDiff) loadTree(oid types.ObjectID) (object.Tree, error) {

	if oid == types.NullObjectID {
		return nil, errors.New("null object id")
	}

	o, err := td.ol.Load(oid.String())
	if err != nil {
		return nil, err
	}

	switch o.Class() {
	case object.ClassCommit:
		c, err := object.ParseCommit(o)
		if err != nil {
			return nil, err
		}
		o, err := td.ol.Load(c.Tree())
		if err != nil {
			return nil, err
		}
		t, err := object.ParseTree(o)
		if err != nil {
			return nil, err
		}
		return t, nil
	case object.ClassTree:
		t, err := object.ParseTree(o)
		if err != nil {
			return nil, err
		}
		return t, nil
	}

	return nil, errors.New("object is not tree")

}

func (td *treeDiff) detectDeletion(a, b map[string]object.TreeEntry, prefix string) error {

	for basename, aChild := range a {

		bChild, ok := b[basename]

		if ok && (aChild.OID() == bChild.OID() && aChild.Permission() == bChild.Permission()) {
			continue
		}

		path := filepath.Join(prefix, basename)

		// tree
		a := types.NullObjectID
		if aChild.IsTree() {
			a = types.ObjectID(aChild.OID())
		}

		b := types.NullObjectID
		if ok && bChild.IsTree() {
			b = types.ObjectID(bChild.OID())
		}

		if err := td.compare(a, b, path); err != nil {
			return err
		}

		// blob
		var c object.TreeEntry = nil
		if !aChild.IsTree() {
			c = aChild
		}

		var o object.TreeEntry = nil
		if ok && !bChild.IsTree() {
			o = bChild
		}

		if c == nil && o == nil {
			continue
		}

		td.changes[path] = newPair(c, o)

	}

	return nil

}

func (td *treeDiff) detectAdditions(a, b map[string]object.TreeEntry, prefix string) error {

	for basename, bChild := range b {

		aChild, ok := a[basename]

		if ok {
			continue
		}

		path := filepath.Join(prefix, basename)
		if bChild.IsTree() {
			td.compare(types.NullObjectID, types.ObjectID(bChild.OID()), path)
		} else {
			td.changes[path] = newPair(aChild, bChild)
		}

	}

	return nil

}

func (td *treeDiff) Changes() map[string]pair {
	return td.changes
}
