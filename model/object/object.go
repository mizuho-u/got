package object

import (
	"crypto/sha1"
	"encoding/hex"
	"strconv"
)

type class string

const (
	classBlob   class = "blob"
	classTree   class = "tree"
	classCommit class = "commit"
)

type Object interface {
	OID() string
	Content() []byte
}

type object struct {
	id      string
	class   class
	content []byte
}

func newObject(data []byte, class class) (*object, error) {

	// object type
	content := []byte(class)

	// a space
	content = append(content, 0x20)

	// the data size in text representation
	content = append(content, []byte(strconv.Itoa(len(data)))...)

	// a null byte
	content = append(content, 0x00)

	// the data
	content = append(content, data...)

	sha1 := sha1.New()
	_, err := sha1.Write(content)
	if err != nil {
		return nil, err
	}

	oid := hex.EncodeToString(sha1.Sum(nil))

	return &object{oid, class, content}, nil
}

func (o *object) OID() string {
	return o.id
}

func (o *object) Content() []byte {
	return o.content
}
