package repository_test

import (
	"testing"

	"github.com/mizuho-u/got/repository"
	"github.com/mizuho-u/got/repository/object"
	"github.com/mizuho-u/got/types"
)

func TestSameFlatTreeDiff(t *testing.T) {

	db := newDatabase()

	as := map[string]*file{
		"hello.txt": {object.RegularFile, []byte("hello")},
		"world.txt": {object.RegularFile, []byte("world")},
	}

	a, objects := newTree(t, as)
	db.store(objects...)

	bs := map[string]*file{
		"hello.txt": {object.RegularFile, []byte("hello")},
		"world.txt": {object.RegularFile, []byte("world")},
	}

	b, objects := newTree(t, bs)
	db.store(objects...)

	diff := repository.NewTreeDiff(db)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	if len(diff.Changes()) != 0 {
		t.Fatalf("expect no changes, but %d changes", len(diff.Changes()))
	}
}

func TestFlatTreeDiff(t *testing.T) {

	db := newDatabase()

	as := map[string]*file{
		"hello.txt":  {object.RegularFile, []byte("hello")},
		"world.txt":  {object.RegularFile, []byte("world")},
		"delete.txt": {object.RegularFile, []byte("delete")},
	}

	a, objects := newTree(t, as)
	db.store(objects...)

	bs := map[string]*file{
		"hello.txt": {object.RegularFile, []byte("hello")},
		"world.txt": {object.RegularFile, []byte("modified")},
		"new.txt":   {object.RegularFile, []byte("new")},
	}

	b, objects := newTree(t, bs)
	db.store(objects...)

	diff := repository.NewTreeDiff(db)
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

	_, ok = changes["world.txt"]
	if !ok {
		t.Error("expect world.txt changed, but unchanged")
	}

	_, ok = changes["delete.txt"]
	if !ok {
		t.Error("expect delete.txt changed, but unchanged")
	}

	_, ok = changes["new.txt"]
	if !ok {
		t.Error("expect new.txt changed, but unchanged")
	}
}

func TestSameNestedTreeDiff(t *testing.T) {

	db := newDatabase()

	as := map[string]*file{
		"hello.txt":     {object.RegularFile, []byte("hello")},
		"world.txt":     {object.RegularFile, []byte("world")},
		"sub/hello.txt": {object.RegularFile, []byte("hello")},
		"sub/world.txt": {object.RegularFile, []byte("world")},
	}

	a, objects := newTree(t, as)
	db.store(objects...)

	bs := map[string]*file{
		"hello.txt":     {object.RegularFile, []byte("hello")},
		"world.txt":     {object.RegularFile, []byte("world")},
		"sub/hello.txt": {object.RegularFile, []byte("hello")},
		"sub/world.txt": {object.RegularFile, []byte("world")},
	}

	b, objects := newTree(t, bs)
	db.store(objects...)

	diff := repository.NewTreeDiff(db)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	if len(diff.Changes()) != 0 {
		t.Fatalf("expect no changes, but %d changes", len(diff.Changes()))
	}
}

func TestNestedTreeDiff(t *testing.T) {

	db := newDatabase()

	as := map[string]*file{
		"hello.txt":      {object.RegularFile, []byte("hello")},
		"world.txt":      {object.RegularFile, []byte("world")},
		"sub/hello.txt":  {object.RegularFile, []byte("hello")},
		"sub/world.txt":  {object.RegularFile, []byte("world")},
		"sub/delete.txt": {object.RegularFile, []byte("delete")},
	}

	a, objects := newTree(t, as)
	db.store(objects...)

	bs := map[string]*file{
		"hello.txt":     {object.RegularFile, []byte("hello")},
		"world.txt":     {object.RegularFile, []byte("modified")},
		"sub/hello.txt": {object.RegularFile, []byte("hello")},
		"sub/world.txt": {object.RegularFile, []byte("modified")},
		"sub/new.txt":   {object.RegularFile, []byte("new")},
	}

	b, objects := newTree(t, bs)
	db.store(objects...)

	diff := repository.NewTreeDiff(db)
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

	if _, ok := changes["sub/hello.txt"]; ok {
		t.Error("expect sub/hello.txt unchanged, but changed")
	}

	if _, ok := changes["world.txt"]; !ok {
		t.Error("expect world.txt changed, but unchanged")
	}

	if _, ok := changes["sub/world.txt"]; !ok {
		t.Error("expect sub/world.txt changed, but unchanged")
	}

	if _, ok := changes["sub/delete.txt"]; !ok {
		t.Error("expect sub/delete.txt deleted, but not deleted")
	}

	if _, ok := changes["sub/new.txt"]; !ok {
		t.Error("expect sub/new.txt created, but not created")
	}
}

func TestTreeDiffReplaceDirectoryWithSameNamedFile(t *testing.T) {

	db := newDatabase()

	as := map[string]*file{
		"hello.txt":     {object.RegularFile, []byte("hello")},
		"world.txt":     {object.RegularFile, []byte("world")},
		"sub/hello.txt": {object.RegularFile, []byte("hello")},
		"sub/world.txt": {object.RegularFile, []byte("world")},
	}

	a, objects := newTree(t, as)
	db.store(objects...)

	bs := map[string]*file{
		"hello.txt": {object.RegularFile, []byte("hello")},
		"world.txt": {object.RegularFile, []byte("world")},
		"sub":       {object.RegularFile, []byte("sub")},
	}

	b, objects := newTree(t, bs)
	db.store(objects...)

	diff := repository.NewTreeDiff(db)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	changes := diff.Changes()
	if len(changes) != 3 {
		t.Fatalf("expect 3 changes, but %d changes", len(changes))
	}

	if _, ok := changes["sub"]; !ok {
		t.Error("expect sub created, but not created")
	}

	if _, ok := changes["sub/hello.txt"]; !ok {
		t.Error("expect sub/hello.txt deleted, but not deleted")
	}

	if _, ok := changes["sub/world.txt"]; !ok {
		t.Error("expect sub/world.txt deleted, but not deleted")
	}

}

func TestTreeDiffChmodFile(t *testing.T) {

	db := newDatabase()

	as := map[string]*file{
		"hello.txt": {object.RegularFile, []byte("hello")},
		"world.txt": {object.RegularFile, []byte("world")},
	}

	a, objects := newTree(t, as)
	db.store(objects...)

	bs := map[string]*file{
		"hello.txt": {object.RegularFile, []byte("hello")},
		"world.txt": {object.ExecutableFile, []byte("world")},
	}

	b, objects := newTree(t, bs)
	db.store(objects...)

	diff := repository.NewTreeDiff(db)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	changes := diff.Changes()
	if len(changes) != 1 {
		t.Fatalf("expect 1 changes, but %d changes", len(changes))
	}

	if _, ok := changes["world.txt"]; !ok {
		t.Error("expect world.txt changed, but not changed")
	}

}

func TestTreeDiffDeleteFile(t *testing.T) {

	db := newDatabase()

	as := map[string]*file{
		"a":     {object.RegularFile, []byte("aaaaa")},
		"b/c":   {object.RegularFile, []byte("ccccc")},
		"b/d/e": {object.RegularFile, []byte("eeeee")},
	}

	a, objects := newTree(t, as)
	db.store(objects...)

	bs := map[string]*file{
		"a":   {object.RegularFile, []byte("aaaaa")},
		"b/c": {object.RegularFile, []byte("ccccc")},
	}

	b, objects := newTree(t, bs)
	db.store(objects...)

	diff := repository.NewTreeDiff(db)
	if err := diff.Diff(types.ObjectID(a.OID()), types.ObjectID(b.OID())); err != nil {
		t.Fatal(err)
	}

	changes := diff.Changes()
	if len(changes) != 1 {
		t.Fatalf("expect 1 changes, but %d changes", len(changes))
	}

	if _, ok := changes["b/d/e"]; !ok {
		t.Error("expect b/d/e.txt deleted, but not deleted")
	}

}
