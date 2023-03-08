package object_test

import (
	"testing"
	"time"

	"github.com/mizuho-u/got/model/object"
)

func TestCreateCommitObject(t *testing.T) {

	tree := "88e38705fdbd3608cddbe904b67c731f3234c45b"

	name := "James Coglan"
	email := "james@jcoglan.com"
	now := time.Unix(1511204319, 0).UTC()

	author := object.NewAuthor(name, email, now)
	commit, err := object.NewCommit("", tree, author.String(), "First commit.\n")
	if err != nil {
		t.Fatal("failed to create commit. ", err)
	}

	if commit.OID() != "2fb7e6b97a594fa7f9ccb927849e95c7c70e39f5" {
		t.Fatalf("commit oid not match. expect %s got %s", "2fb7e6b97a594fa7f9ccb927849e95c7c70e39f5", commit.OID())
	}

}
