package object

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"strings"
)

type class string

const (
	ClassBlob   class = "blob"
	ClassTree   class = "tree"
	ClassCommit class = "commit"
)

type Object interface {
	OID() string
	Content() []byte
	Class() class
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

func ParseObject(rawdata []byte) (Object, error) {

	object := &object{}

	buffer := bytes.NewBuffer(rawdata)

	c, err := buffer.ReadString(0x20)
	if err != nil {
		return nil, err
	}
	object.class = class(strings.TrimSpace(c))

	bytes, err := buffer.ReadBytes(0x00)
	if err != nil {
		return nil, err
	}

	size, err := strconv.Atoi(string(bytes[0 : len(bytes)-1]))
	if err != nil {
		return nil, err
	}

	object.content = buffer.Next(size)

	sha1 := sha1.New()
	_, err = sha1.Write(object.content)
	if err != nil {
		return nil, err
	}

	object.id = hex.EncodeToString(sha1.Sum(nil))

	return object, nil

}

func (o *object) OID() string {
	return o.id
}

func (o *object) Content() []byte {
	return o.content
}

func (o *object) Class() class {
	return o.class
}
