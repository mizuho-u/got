package object

import (
	"bytes"
	"fmt"

	"strings"
)

type Commit interface {
	Object
	Tree() string
	Parent() string
	TitleLine() string
}

type commit struct {
	parent, tree, message string
	author                *author
	*object
}

func NewCommit(parent, tree string, author *author, message string) (*commit, error) {

	content := []byte{}

	content = append(content, []byte("tree "+tree+"\n")...)
	if parent != "" {
		content = append(content, []byte("parent "+parent+"\n")...)
	}
	content = append(content, []byte("author "+author.String()+"\n")...)
	content = append(content, []byte("committer "+author.String()+"\n")...)
	content = append(content, []byte("\n")...)
	content = append(content, []byte(message)...)

	object, err := newObject(content, ClassCommit)
	if err != nil {
		return nil, err
	}

	return &commit{parent, tree, message, author, object}, nil
}

func EmptyCommit() Commit {

	c := &commit{object: &object{id: ""}}
	return c
}

func ParseCommit(obj Object) (Commit, error) {

	if obj.Class() != ClassCommit {
		return nil, fmt.Errorf("object is not commit: %s", obj.Class())
	}

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

	author, err := authorFromString(strings.TrimRight(strings.TrimLeft(str, "author "), "\n"))
	if err != nil {
		return nil, err
	}
	c.author = author

	str, err = buf.ReadString(0x0A)
	if err != nil {
		return nil, err
	}
	committer, err := authorFromString(strings.TrimSuffix(strings.TrimPrefix(str, "committer "), "\n"))
	if err != nil {
		return nil, err
	}
	c.author = committer

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

func (c *commit) Parent() string {
	return c.parent
}

func (c *commit) TitleLine() string {
	return fmt.Sprintf("%s - %s", c.author.now.Format("2006-01-02"), strings.Split(c.message, "\n")[0])
}
