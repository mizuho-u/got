package repository_test

import (
	"testing"

	"github.com/mizuho-u/got/repository"
	"github.com/mizuho-u/got/repository/object"
	"github.com/mizuho-u/got/types"
)

func TestMigration(t *testing.T) {

	testt := []struct {
		description string
		a           map[string]*file
		b           map[string]*file
	}{
		{
			description: "same tree contents",
			a: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("eeeee")},
			},
			b: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("eeeee")},
			},
		},
		{
			description: "update files",
			a: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("eeeee")},
			},
			b: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("modified")},
				"b/d/e": {object.ExecutableFile, []byte("eeeee")},
			},
		},
		{
			description: "some new files",
			a: map[string]*file{
				"a":   {object.RegularFile, []byte("aaaaa")},
				"b/c": {object.RegularFile, []byte("ccccc")},
			},
			b: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("new")},
			},
		},
		{
			description: "all new files",
			a:           map[string]*file{},
			b: map[string]*file{
				"a":     {object.RegularFile, []byte("new")},
				"b/c":   {object.RegularFile, []byte("new")},
				"b/d/e": {object.RegularFile, []byte("new")},
			},
		},
		{
			description: "delete some files",
			a: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("eeeee")},
				"b/d/f": {object.RegularFile, []byte("fffff")},
			},
			b: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/f": {object.RegularFile, []byte("fffff")},
			},
		},
		{
			description: "delete some files and empty directories",
			a: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("eeeee")},
				"b/d/f": {object.RegularFile, []byte("fffff")},
			},
			b: map[string]*file{
				"a":   {object.RegularFile, []byte("aaaaa")},
				"b/c": {object.RegularFile, []byte("ccccc")},
			},
		},
		{
			description: "delete all file",
			a: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("eeeee")},
			},
			b: map[string]*file{},
		},
		{
			description: "change directory files",
			a: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("eeeee")},
			},
			b: map[string]*file{
				"a": {object.RegularFile, []byte("aaaaa")},
				"b": {object.RegularFile, []byte("bbbbb")},
			},
		},
		{
			description: "change, new, delete files",
			a: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("ccccc")},
				"b/d/e": {object.RegularFile, []byte("deleted")},
			},
			b: map[string]*file{
				"a":     {object.RegularFile, []byte("aaaaa")},
				"b/c":   {object.RegularFile, []byte("modified")},
				"b/d/f": {object.RegularFile, []byte("new")},
			},
		},
	}

	for _, tc := range testt {

		t.Run(tc.description, func(t *testing.T) {

			ws := newWorkspace()
			db := newDatabase()

			ws.addRange(tc.a)

			atree, objects := newTree(t, tc.a)
			db.store(objects...)

			btree, objects := newTree(t, tc.b)
			db.store(objects...)

			diff := repository.NewTreeDiff(db)
			if err := diff.Diff(types.ObjectID(atree.OID()), types.ObjectID(btree.OID())); err != nil {
				t.Fatal(err)
			}

			m := repository.NewMigration(diff.Changes(), ws, db)
			if err := m.ApplyChanges(); err != nil {
				t.Fatal(err)
			}

			ws.equals(t, tc.b)

		})

	}

}
