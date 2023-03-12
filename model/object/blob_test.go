package object_test

import (
	"testing"

	"github.com/mizuho-u/got/model/object"
)

func TestCreateBlobObject(t *testing.T) {

	tests := []struct {
		data     []byte
		filename string
		oid      string
	}{
		{data: []byte("hello\n"), filename: "hello.txt", oid: "ce013625030ba8dba906f756967f9e9ca394464a"},
	}

	for _, tt := range tests {

		blob, err := object.NewBlob(tt.filename, tt.data)
		if err != nil {
			t.Error("failed to create blob ", err)
			continue
		}

		if blob.OID() != tt.oid {
			t.Errorf("unexpected oid. expect %s got %s", tt.oid, blob.OID())
		}

	}

}
