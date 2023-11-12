package object

import (
	"bytes"

	"strings"
)

type Commit interface {
	Object
	Tree() string
}

type commit struct {
	parent, tree, author, message string
	*object
}

func NewCommit(parent, tree, author, message string) (*commit, error) {

	content := []byte{}

	content = append(content, []byte("tree "+tree+"\n")...)
	if parent != "" {
		content = append(content, []byte("parent "+parent+"\n")...)
	}
	content = append(content, []byte("author "+author+"\n")...)
	content = append(content, []byte("committer "+author+"\n")...)
	content = append(content, []byte("\n")...)
	content = append(content, []byte(message)...)

	object, err := newObject(content, ClassCommit)
	if err != nil {
		return nil, err
	}

	return &commit{parent, tree, author, message, object}, nil
}

func EmptyCommit() Commit {

	c := &commit{object: &object{id: ""}}
	return c
}

func ParseCommit(obj Object) (Commit, error) {

	c := &commit{object: &object{id: obj.OID()}}

	buf := bytes.NewBuffer(obj.Data())

	str, err := buf.ReadString(0x0A)
	if err != nil {
		return nil, err
	}
	c.tree = strings.TrimSuffix(strings.TrimPrefix(str, "tree "), "\n")

	str, err = buf.ReadString(0x0A)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(str, "parent") {

		c.parent = strings.TrimRight(strings.TrimLeft(str, "parent "), "\n")

		str, err = buf.ReadString(0x0A)
		if err != nil {
			return nil, err
		}
	}

	c.author = strings.TrimRight(strings.TrimLeft(str, "author "), "\n")

	str, err = buf.ReadString(0x0A)
	if err != nil {
		return nil, err
	}
	c.author = strings.TrimSuffix(strings.TrimPrefix(str, "committer "), "\n")

	_, err = buf.ReadByte()
	if err != nil {
		return nil, err
	}

	c.message = buf.String()

	return c, nil

}

func (c *commit) Tree() string {
	return c.tree
}
