package model_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/mizuho-u/got/model"
)

func TestCommitWorkspace(t *testing.T) {

	// arrange
	// act
	ws, err := model.NewWorkspace()
	if err != nil {
		t.Fatal("create workspace failed. ", err)
	}

	commitId, err := ws.Commit("", "Mizuho Ueda", "mi_ueda@u-m.dev", "First Commit.", getTimeInJst(t, 1511204319), &model.File{Name: "hello.txt", Data: []byte("Hello"), Permission: 0644})
	if err != nil {
		t.Fatal("commit failed. ", err)
	}

	// assert
	if commitId != "a5969546fc417f4b362e5290ad8ee49b044bfc0e" {
		t.Fatalf("commitId not match. expect %s got %s", "a5969546fc417f4b362e5290ad8ee49b044bfc0e", commitId)
	}

	if len(ws.Objects()) != 3 {
		t.Fatalf("unexpected objects length want 3 got %d", len(ws.Objects()))
	}

	// expect every object class to be created
	created := 0b0000
	for _, o := range ws.Objects() {

		switch getclass(o.Content()) {
		case "blob":
			created |= 0b0001
		case "tree":
			created |= 0b0010
		case "commit":
			created |= 0b0100
		}

	}

	if created != 0b0111 {
		t.Fatalf("missing blob, tree or commit objects %b", created)
	}

}

func getTimeInJst(t *testing.T, unixsec int64) time.Time {

	t.Helper()

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}

	return time.Unix(unixsec, 0).In(jst)

}

func getclass(content []byte) string {

	class := []byte{}

	for _, b := range content {
		if b == 0x20 {
			break
		}
		class = append(class, b)
	}

	return string(class)

}

func TestNewWorkspaceWithIndex(t *testing.T) {

	source := []byte{
		0x44, 0x49, 0x52, 0x43, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x64, 0x07, 0x3e, 0x95,
		0x02, 0xe7, 0x75, 0xbf, 0x64, 0x07, 0x03, 0xa2, 0x20, 0x0a, 0x70, 0x35, 0x01, 0x00, 0x00, 0x11,
		0x03, 0x84, 0x45, 0x6a, 0x00, 0x00, 0x81, 0xa4, 0x00, 0x00, 0x01, 0xf5, 0x00, 0x00, 0x00, 0x14,
		0x00, 0x00, 0x05, 0x0c, 0x91, 0x98, 0x9b, 0xfa, 0xee, 0x2e, 0x41, 0xbe, 0x1e, 0x9d, 0x30, 0x81,
		0xeb, 0x3d, 0x39, 0x06, 0x21, 0x4e, 0x2e, 0x03, 0x00, 0x0a, 0x63, 0x6d, 0x64, 0x2f, 0x61, 0x64,
		0x64, 0x2e, 0x67, 0x6f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf3, 0x5c, 0xeb, 0x94,
		0xef, 0x07, 0xbf, 0xac, 0x40, 0xfb, 0x34, 0x1e, 0x19, 0x88, 0x6e, 0x05, 0x96, 0x94, 0x5e, 0x06,
	}

	ws, err := model.NewWorkspace(model.WithIndex(bytes.NewBuffer(source)))
	if err != nil {
		t.Fatal(err)
	}

	serialized, err := ws.Index().Serialize()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(source, serialized) {
		t.Fatal("serialized index not match to source")
	}

}
