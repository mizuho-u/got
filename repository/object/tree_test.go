package object_test

import (
	"testing"

	"github.com/mizuho-u/got/repository/object"
)

func TestFlatTree(t *testing.T) {

	entries := []object.TreeEntry{}

	entries = append(entries, object.NewTreeEntry("world.txt", object.RegularFile, "cc628ccd10742baea8241c5924df992b5c019f71"))
	entries = append(entries, object.NewTreeEntry("hello.txt", object.RegularFile, "ce013625030ba8dba906f756967f9e9ca394464a"))

	tree, err := object.BuildTree(entries)
	if err != nil {
		t.Fatal("failed to create tree. ", err)
	}

	if tree.OID() != "88e38705fdbd3608cddbe904b67c731f3234c45b" {
		t.Fatalf("tree oid not match. want %s, got %s", "88e38705fdbd3608cddbe904b67c731f3234c45b", tree.OID())
	}

}

func TestNestedTree(t *testing.T) {

	entries := []object.TreeEntry{}

	entries = append(entries, object.NewTreeEntry("a.txt", object.RegularFile, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"))
	entries = append(entries, object.NewTreeEntry("b/d/e.txt", object.RegularFile, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"))
	entries = append(entries, object.NewTreeEntry("b/h.txt", object.RegularFile, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"))
	entries = append(entries, object.NewTreeEntry("c.txt", object.RegularFile, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"))
	entries = append(entries, object.NewTreeEntry("f/g.txt", object.RegularFile, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"))

	tree, err := object.BuildTree(entries)
	if err != nil {
		t.Fatal("failed to create tree. ", err)
	}

	if tree.OID() != "a2a6a1a5f6f0ada6870b93baac50ecc7e5cb6f03" {
		t.Fatalf("tree oid not match. want %s, got %s", "a2a6a1a5f6f0ada6870b93baac50ecc7e5cb6f03", tree.OID())
	}

}

func TestTreeEntriesOrder(t *testing.T) {

	entries := []object.TreeEntry{}

	entries = append(entries, object.NewTreeEntry("foo.txt", object.RegularFile, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"))
	entries = append(entries, object.NewTreeEntry("foo/bar.txt", object.RegularFile, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"))

	tree, err := object.BuildTree(entries)
	if err != nil {
		t.Fatal("failed to create tree. ", err)
	}

	if tree.OID() != "d811d36dfed04b19f4b009cae7a1a7a92c7f1918" {
		t.Fatalf("tree oid not match. want %s, got %s", "d811d36dfed04b19f4b009cae7a1a7a92c7f1918", tree.OID())
	}

}
