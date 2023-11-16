package repository_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mizuho-u/got/repository"
	"github.com/mizuho-u/got/repository/object"
	"github.com/mizuho-u/got/types"
)

type ol struct {
	objects map[string]object.Object
}

func (ol *ol) Load(oid string) (object.Object, error) {

	if o, ok := ol.objects[oid]; ok {
		return o, nil
	} else {
		return nil, fmt.Errorf("object not found. id: %s", oid)
	}

}

func (ol *ol) Add(os ...object.Object) {

	for _, o := range os {
		ol.objects[o.OID()] = o
	}

}

func TestSameFlatTreeDiff(t *testing.T) {

	aHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	aWorld, _ := object.NewBlob("world.txt", []byte("world"))

	aEntries := []object.TreeEntry{}
	aEntries = append(aEntries, object.NewTreeEntry(aHello.Filename(), object.RegularFile, aHello.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(aWorld.Filename(), object.RegularFile, aWorld.OID()))

	a, err := object.BuildTree(aEntries)
	if err != nil {
		t.Error(err)
	}

	bHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	bWorld, _ := object.NewBlob("world.txt", []byte("world"))

	bEntries := []object.TreeEntry{}
	bEntries = append(bEntries, object.NewTreeEntry(bHello.Filename(), object.RegularFile, bHello.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(bWorld.Filename(), object.RegularFile, bWorld.OID()))

	b, err := object.BuildTree(bEntries)
	if err != nil {
		t.Error(err)
	}

	ol := &ol{map[string]object.Object{}}
	ol.Add(a, b)

	diff := repository.NewTreeDiff(ol)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	if len(diff.Changes()) != 0 {
		t.Fatalf("expect no changes, but %d changes", len(diff.Changes()))
	}
}

func TestFlatTreeDiff(t *testing.T) {

	aHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	aWorld, _ := object.NewBlob("world.txt", []byte("world"))
	aDelete, _ := object.NewBlob("delete.txt", []byte("delete"))

	aEntries := []object.TreeEntry{}
	aEntries = append(aEntries, object.NewTreeEntry(aHello.Filename(), object.RegularFile, aHello.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(aWorld.Filename(), object.RegularFile, aWorld.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(aDelete.Filename(), object.RegularFile, aDelete.OID()))

	a, err := object.BuildTree(aEntries)
	if err != nil {
		t.Error(err)
	}

	bHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	bWorld, _ := object.NewBlob("world.txt", []byte("modified"))
	bNew, _ := object.NewBlob("new.txt", []byte("new"))

	bEntries := []object.TreeEntry{}
	bEntries = append(bEntries, object.NewTreeEntry(bHello.Filename(), object.RegularFile, bHello.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(bWorld.Filename(), object.RegularFile, bWorld.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(bNew.Filename(), object.RegularFile, bNew.OID()))

	b, err := object.BuildTree(bEntries)
	if err != nil {
		t.Error(err)
	}

	ol := &ol{map[string]object.Object{}}
	ol.Add(a, b)

	diff := repository.NewTreeDiff(ol)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	changes := diff.Changes()
	if len(changes) != 3 {
		t.Fatalf("expect 3 changes, but %d changes", len(changes))
	}

	_, ok := changes["hello.txt"]
	if ok {
		t.Error("expect hello.txt unchanged, but changed")
	}

	modified := changes["world.txt"]
	if a, b := modified.Item1().OID() == aWorld.OID(), modified.Item2().OID() == bWorld.OID(); !(a && b) {
		t.Errorf("modified file not match a %t b %t", a, b)
	}

	deleted := changes["delete.txt"]
	if a, b := deleted.Item1().OID() == aDelete.OID(), deleted.Item2() == nil; !(a && b) {
		t.Errorf("deleted file not match a %t b %t", a, b)
	}

	added := changes["new.txt"]
	if a, b := added.Item1() == nil, added.Item2().OID() == bNew.OID(); !(a && b) {
		t.Errorf("added file not match a %t b %t", a, b)
	}

}

func TestSameNestedTreeDiff(t *testing.T) {

	// a
	aSubHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	aSubWorld, _ := object.NewBlob("world.txt", []byte("world"))

	aSubEntries := []object.TreeEntry{}
	aSubEntries = append(aSubEntries, object.NewTreeEntry(aSubHello.Filename(), object.RegularFile, aSubHello.OID()))
	aSubEntries = append(aSubEntries, object.NewTreeEntry(aSubWorld.Filename(), object.RegularFile, aSubWorld.OID()))

	aSub, err := object.BuildTree(aSubEntries)
	if err != nil {
		t.Fatal(err)
	}

	aEntries := []object.TreeEntry{}
	aHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	aWorld, _ := object.NewBlob("world.txt", []byte("world"))

	aEntries = append(aEntries, object.NewTreeEntry(aHello.Filename(), object.RegularFile, aHello.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(aWorld.Filename(), object.RegularFile, aWorld.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(filepath.Join("sub", aSubHello.Filename()), object.RegularFile, aSubHello.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(filepath.Join("sub", aSubWorld.Filename()), object.RegularFile, aSubWorld.OID()))

	a, err := object.BuildTree(aEntries)
	if err != nil {
		t.Fatal(err)
	}

	// b
	bSubHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	bSubWorld, _ := object.NewBlob("world.txt", []byte("world"))

	bSubEntries := []object.TreeEntry{}
	bSubEntries = append(bSubEntries, object.NewTreeEntry(bSubHello.Filename(), object.RegularFile, bSubHello.OID()))
	bSubEntries = append(bSubEntries, object.NewTreeEntry(bSubWorld.Filename(), object.RegularFile, bSubWorld.OID()))

	bSub, err := object.BuildTree(bSubEntries)
	if err != nil {
		t.Fatal(err)
	}

	bHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	bWorld, _ := object.NewBlob("world.txt", []byte("world"))

	bEntries := []object.TreeEntry{}
	bEntries = append(bEntries, object.NewTreeEntry(bHello.Filename(), object.RegularFile, bHello.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(bWorld.Filename(), object.RegularFile, bWorld.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(filepath.Join("sub", bSubHello.Filename()), object.RegularFile, bSubHello.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(filepath.Join("sub", bSubWorld.Filename()), object.RegularFile, bSubWorld.OID()))

	b, err := object.BuildTree(bEntries)
	if err != nil {
		t.Fatal(err)
	}

	ol := &ol{map[string]object.Object{}}
	ol.Add(a, b, aSub, bSub)

	diff := repository.NewTreeDiff(ol)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	if len(diff.Changes()) != 0 {
		t.Fatalf("expect no changes, but %d changes", len(diff.Changes()))
	}
}

func TestNestedTreeDiff(t *testing.T) {

	// a
	aSubHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	aSubWorld, _ := object.NewBlob("world.txt", []byte("world"))
	aSubDelete, _ := object.NewBlob("delete.txt", []byte("delete"))

	aSubEntries := []object.TreeEntry{}
	aSubEntries = append(aSubEntries, object.NewTreeEntry(aSubHello.Filename(), object.RegularFile, aSubHello.OID()))
	aSubEntries = append(aSubEntries, object.NewTreeEntry(aSubWorld.Filename(), object.RegularFile, aSubWorld.OID()))
	aSubEntries = append(aSubEntries, object.NewTreeEntry(aSubDelete.Filename(), object.RegularFile, aSubDelete.OID()))

	aSub, err := object.BuildTree(aSubEntries)
	if err != nil {
		t.Fatal(err)
	}

	aHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	aWorld, _ := object.NewBlob("world.txt", []byte("world"))

	aEntries := []object.TreeEntry{}
	aEntries = append(aEntries, object.NewTreeEntry(filepath.Join("sub", aSubHello.Filename()), object.RegularFile, aSubHello.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(filepath.Join("sub", aSubWorld.Filename()), object.RegularFile, aSubWorld.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(filepath.Join("sub", aSubDelete.Filename()), object.RegularFile, aSubDelete.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(aHello.Filename(), object.RegularFile, aHello.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(aWorld.Filename(), object.RegularFile, aWorld.OID()))

	a, err := object.BuildTree(aEntries)
	if err != nil {
		t.Fatal(err)
	}

	// b
	bSubHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	bSubWorld, _ := object.NewBlob("world.txt", []byte("modified"))
	bSubNew, _ := object.NewBlob("new.txt", []byte("new"))

	bSubEntries := []object.TreeEntry{}
	bSubEntries = append(bSubEntries, object.NewTreeEntry(bSubHello.Filename(), object.RegularFile, bSubHello.OID()))
	bSubEntries = append(bSubEntries, object.NewTreeEntry(bSubWorld.Filename(), object.RegularFile, bSubWorld.OID()))
	bSubEntries = append(bSubEntries, object.NewTreeEntry(bSubNew.Filename(), object.RegularFile, bSubNew.OID()))

	bSub, err := object.BuildTree(bSubEntries)
	if err != nil {
		t.Fatal(err)
	}

	bHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	bWorld, _ := object.NewBlob("world.txt", []byte("modified"))

	bEntries := []object.TreeEntry{}
	bEntries = append(bEntries, object.NewTreeEntry(filepath.Join("sub", bSubHello.Filename()), object.RegularFile, bSubHello.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(filepath.Join("sub", bSubWorld.Filename()), object.RegularFile, bSubWorld.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(filepath.Join("sub", bSubNew.Filename()), object.RegularFile, bSubNew.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(bHello.Filename(), object.RegularFile, bHello.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(bWorld.Filename(), object.RegularFile, bWorld.OID()))

	b, err := object.BuildTree(bEntries)
	if err != nil {
		t.Fatal(err)
	}

	ol := &ol{map[string]object.Object{}}
	ol.Add(a, b, aSub, bSub)

	diff := repository.NewTreeDiff(ol)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	changes := diff.Changes()

	if len(changes) != 4 {
		t.Fatalf("expect 4 changes, but %d changes", len(changes))
	}

	if _, ok := changes["hello.txt"]; ok {
		t.Error("expect hello.txt unchanged, but changed")
	}

	modified := changes["world.txt"]
	if a, b := modified.Item1().OID() == aWorld.OID(), modified.Item2().OID() == bWorld.OID(); !(a && b) {
		t.Errorf("modified file not match a %t b %t", a, b)
	}

	if _, ok := changes["sub/hello.txt"]; ok {
		t.Error("expect sub/hello.txt unchanged, but changed")
	}

	modified = changes["sub/world.txt"]
	if a, b := modified.Item1().OID() == aSubWorld.OID(), modified.Item2().OID() == bSubWorld.OID(); !(a && b) {
		t.Errorf("modified file not match a %t b %t", a, b)
	}

	deleted := changes["sub/delete.txt"]
	if a, b := deleted.Item1().OID() == aSubDelete.OID(), deleted.Item2() == nil; !(a && b) {
		t.Errorf("deleted file not match a %t b %t", a, b)
	}

	added := changes["sub/new.txt"]
	if a, b := added.Item1() == nil, added.Item2().OID() == bSubNew.OID(); !(a && b) {
		t.Errorf("added file not match a %t b %t", a, b)
	}

}

func TestTreeDiffReplaceDirectoryWithSameNamedFile(t *testing.T) {

	// a
	aaHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	aaWorld, _ := object.NewBlob("world.txt", []byte("world"))

	aaEntries := []object.TreeEntry{}
	aaEntries = append(aaEntries, object.NewTreeEntry(aaHello.Filename(), object.RegularFile, aaHello.OID()))
	aaEntries = append(aaEntries, object.NewTreeEntry(aaWorld.Filename(), object.RegularFile, aaWorld.OID()))

	aa, err := object.BuildTree(aaEntries)
	if err != nil {
		t.Fatal(err)
	}

	aHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	aWorld, _ := object.NewBlob("world.txt", []byte("world"))

	aEntries := []object.TreeEntry{}
	aEntries = append(aEntries, object.NewTreeEntry(filepath.Join("sub", aaHello.Filename()), object.RegularFile, aaHello.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(filepath.Join("sub", aaWorld.Filename()), object.RegularFile, aaWorld.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(aHello.Filename(), object.RegularFile, aHello.OID()))
	aEntries = append(aEntries, object.NewTreeEntry(aWorld.Filename(), object.RegularFile, aWorld.OID()))

	a, err := object.BuildTree(aEntries)
	if err != nil {
		t.Fatal(err)
	}

	// b
	bHello, _ := object.NewBlob("hello.txt", []byte("hello"))
	bWorld, _ := object.NewBlob("world.txt", []byte("world"))
	bSub, _ := object.NewBlob("sub", []byte("sub"))

	bEntries := []object.TreeEntry{}
	bEntries = append(bEntries, object.NewTreeEntry(bHello.Filename(), object.RegularFile, bHello.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(bWorld.Filename(), object.RegularFile, bWorld.OID()))
	bEntries = append(bEntries, object.NewTreeEntry(bSub.Filename(), object.RegularFile, bSub.OID()))

	b, err := object.BuildTree(bEntries)
	if err != nil {
		t.Fatal(err)
	}

	ol := &ol{map[string]object.Object{}}
	ol.Add(a, b, aa)

	diff := repository.NewTreeDiff(ol)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	changes := diff.Changes()
	if len(changes) != 3 {
		t.Fatalf("expect 3 changes, but %d changes", len(changes))
	}

	added := changes["sub"]
	if a, b := added.Item1() == nil, added.Item2().OID() == bSub.OID(); !(a && b) {
		t.Errorf("added file not match a %t b %t", a, b)
	}

	deleted := changes["sub/hello.txt"]
	if a, b := deleted.Item1().OID() == aaHello.OID(), deleted.Item2() == nil; !(a && b) {
		t.Errorf("deleted file not match a %t b %t", a, b)
	}

	deleted = changes["sub/world.txt"]
	if a, b := deleted.Item1().OID() == aaWorld.OID(), deleted.Item2() == nil; !(a && b) {
		t.Errorf("deleted file not match a %t b %t", a, b)
	}

}
