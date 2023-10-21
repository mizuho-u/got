package object

import (
	"bytes"
	"strings"
)

type Commit struct {
	parent, tree, author, message string
	*object
}

func NewCommit(parent, tree, author, message string) (*Commit, error) {

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

	return &Commit{parent, tree, author, message, object}, nil
}

func ParseCommit(object Object) (*Commit, error) {

	commit := &Commit{}

	buf := bytes.NewBuffer(object.Content())

	str, err := buf.ReadString(0x0A)
	if err != nil {
		return nil, err
	}
	commit.tree = strings.TrimSuffix(strings.TrimPrefix(str, "tree "), "\n")

	str, err = buf.ReadString(0x0A)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(str, "parent") {

		commit.parent = strings.TrimRight(strings.TrimLeft(str, "parent "), "\n")

		str, err = buf.ReadString(0x0A)
		if err != nil {
			return nil, err
		}
	}

	commit.author = strings.TrimRight(strings.TrimLeft(str, "author "), "\n")

	str, err = buf.ReadString(0x0A)
	if err != nil {
		return nil, err
	}
	commit.author = strings.TrimSuffix(strings.TrimPrefix(str, "committer "), "\n")

	_, err = buf.ReadByte()
	if err != nil {
		return nil, err
	}

	commit.message = buf.String()

	return commit, nil

}

func (c *Commit) Tree() string {
	return c.tree
}
