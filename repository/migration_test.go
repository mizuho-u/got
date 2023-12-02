package repository_test

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
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
			index, _ := repository.NewIndex()

			ws.addRange(tc.a)

			atree, objects := newTree(t, tc.a)
			db.store(objects...)

			btree, objects := newTree(t, tc.b)
			db.store(objects...)

			diff := repository.NewTreeDiff(db)
			if err := diff.Diff(types.ObjectID(atree.OID()), types.ObjectID(btree.OID())); err != nil {
				t.Fatal(err)
			}

			m := repository.NewMigration(diff.Changes(), ws, db, index, repository.NewInspector(index, ws))
			if err := m.ApplyChanges(); err != nil {
				t.Fatal(err)
			}

			ws.equals(t, tc.b)

		})

	}

}

func TestMigrationConflict(t *testing.T) {

	testt := []struct {
		description string
		a           map[string]*file
		b           map[string]*file
		modify      map[string]*file
		add         []string
		conflicts   []string
	}{
		{
			description: "unstaged stale file",
			a: map[string]*file{
				"a": {object.RegularFile, []byte("aaaaa")},
			},
			b: map[string]*file{
				"a": {object.RegularFile, []byte("bbbbb")},
			},
			modify: map[string]*file{
				"a": {object.RegularFile, []byte("ccccc")},
			},
			conflicts: []string{`Your local changes to the following files would be overwritten by checkout:
	a
Please commit your changes or stash them before you switch branches.`,
			},
		},
		{
			description: "staged stale file",
			a: map[string]*file{
				"a": {object.RegularFile, []byte("aaaaa")},
			},
			b: map[string]*file{
				"a": {object.RegularFile, []byte("bbbbb")},
			},
			modify: map[string]*file{
				"a": {object.RegularFile, []byte("ccccc")},
			},
			add: []string{"a"},
			conflicts: []string{`Your local changes to the following files would be overwritten by checkout:
	a
Please commit your changes or stash them before you switch branches.`,
			},
		},
		{
			description: "stale directory",
			a: map[string]*file{
				"a": {object.RegularFile, []byte("aaaaa")},
			},
			b: map[string]*file{
				"b": {object.RegularFile, []byte("bbbbb")},
			},
			modify: map[string]*file{
				"b/c": {object.RegularFile, []byte("ccccc")},
				"b/d": {object.RegularFile, []byte("ddddd")},
			},
			add: []string{"b/d"},
			conflicts: []string{`Updating the following directoris would lose untracked files in them:
	b

`,
			},
		},
		{
			description: "untracked file overwritten",
			a: map[string]*file{
				"a": {object.RegularFile, []byte("aaaaa")},
			},
			b: map[string]*file{
				"b/c": {object.RegularFile, []byte("bbbbb")},
			},
			modify: map[string]*file{
				"b": {object.RegularFile, []byte("ccccc")},
			},
			conflicts: []string{`The following untracked working tree files would be overwritten by checkout:
	b
Please move or remove them before you switch branches`,
			},
		},
		// 		{
		// 			description: "untracked file removal",
		// 			a: map[string]*file{
		// 				"a": {object.RegularFile, []byte("aaaaa")},
		// 			},
		// 			b: map[string]*file{
		// 				"b/c": {object.RegularFile, []byte("bbbbb")},
		// 			},
		// 			modify: map[string]*file{
		// 				"b": {object.RegularFile, []byte("ccccc")},
		// 			},
		// 			conflicts: []string{`The following untracked working tree files would be removed by checkout:
		// 	b
		// Please move or remove them before you switch branches`,
		// 			},
		// 		},
	}

	for _, tc := range testt {

		t.Run(tc.description, func(t *testing.T) {

			ws := newWorkspace()
			db := newDatabase()
			index, _ := repository.NewIndex()

			ws.addRange(tc.a)

			// old
			atree, objects := newTree(t, tc.a)
			db.store(objects...)

			for path, data := range tc.a {

				blob, _ := object.NewBlob(path, data.data)

				stat, _ := ws.Stat(path)

				index.Add(repository.NewIndexEntry(path, blob.OID(), stat.Stats()))
			}

			// modify
			ws.modify(tc.modify)

			for _, path := range tc.add {

				f, err := ws.Open(path)
				if err != nil {
					t.Fatal(err)
				}

				data, err := io.ReadAll(f)
				if err != nil {
					t.Fatal(err)
				}

				blob, _ := object.NewBlob(path, data)

				index.Add(repository.NewIndexEntry(path, blob.OID(), f.Info().Stats()))
			}

			// checkout
			btree, objects := newTree(t, tc.b)
			db.store(objects...)

			diff := repository.NewTreeDiff(db)
			if err := diff.Diff(types.ObjectID(atree.OID()), types.ObjectID(btree.OID())); err != nil {
				t.Fatal(err)
			}

			m := repository.NewMigration(diff.Changes(), ws, db, index, repository.NewInspector(index, ws))
			if err := m.ApplyChanges(); err != nil {
				t.Fatal(err)
			}

			conflicts := m.Conflicts()
			if len(conflicts) != len(tc.conflicts) {
				t.Fatalf("expect %d conflicts, got %d conflicts", len(tc.conflicts), len(conflicts))
			}

			for i, conflict := range conflicts {

				if diff := cmp.Diff(conflict.Error(), tc.conflicts[i]); diff != "" {
					t.Errorf("conflict diff %s", diff)
				}

			}

		})

	}

}
