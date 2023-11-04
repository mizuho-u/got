package object

import (
	"bytes"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mizuho-u/got/model/internal"
)

type Permission string

const (
	RegularFile    Permission = "100644"
	ExecutableFile Permission = "100755"
	Directory      Permission = "40000"
)

type treeEntry struct {
	filepath string
	oid      string
	perm     Permission
}

func (te *treeEntry) OID() string {
	return te.oid
}

func (te *treeEntry) Basename() string {
	return filepath.Base(te.filepath)
}

func (te *treeEntry) fullpath() string {
	return te.filepath
}

func (te *treeEntry) Permission() Permission {
	return te.perm
}

func (te *treeEntry) build() error {
	return nil
}

func (te *treeEntry) IsTree() bool {
	return false
}

type TreeEntry interface {
	OID() string
	IsTree() bool
	Basename() string
	fullpath() string
	Permission() Permission
	build() error
}

func NewTreeEntry(filepath string, permission Permission, oid string) TreeEntry {
	return &treeEntry{filepath: filepath, perm: permission, oid: oid}
}

type tree struct {
	*object
	full     string
	base     string
	children []TreeEntry
	index    map[string]int
}

func BuildTree(entries []TreeEntry) (*tree, error) {

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].fullpath() < entries[j].fullpath()
	})

	root := &tree{children: []TreeEntry{}, index: map[string]int{}}
	for _, e := range entries {
		root.add(internal.ParentDirs(e.fullpath()), e)
	}

	if err := root.build(); err != nil {
		return nil, err
	}

	return root, nil
}

func ParseTree(o Object) (*tree, error) {

	root := &tree{children: []TreeEntry{}}

	buf := bytes.NewBuffer(o.Data())

	for buf.Len() > 0 {

		perm, err := buf.ReadString(0x20)
		if err != nil {
			return nil, err
		}
		perm = strings.TrimSpace(perm)

		bytes, err := buf.ReadBytes(0x00)
		if err != nil {
			return nil, err
		}
		filepath := string(bytes[0 : len(bytes)-1])

		oid := internal.Unpack(buf.Next(20))

		if p := Permission(perm); p == Directory {
			root.children = append(root.children, &tree{children: []TreeEntry{}, base: filepath, object: &object{id: oid}})
		} else {
			root.children = append(root.children, &treeEntry{perm: p, filepath: filepath, oid: oid})
		}

	}

	return root, nil

}

func (t *tree) add(parents []string, e TreeEntry) {

	if len(parents) == 0 {
		t.index[e.Basename()] = len(t.children)
		t.children = append(t.children, e)

	} else {

		base := filepath.Base(parents[0])

		ct, ok := t.getChildTree(base)
		if !ok {
			t.addChildTree(base)
			ct, _ = t.getChildTree(base)
		}

		ct.add(parents[1:], e)
	}

}

func (t *tree) getChildTree(basepath string) (*tree, bool) {

	index, ok := t.index[basepath]
	if !ok {
		return nil, false
	}

	ct, ok := t.children[index].(*tree)
	if !ok {
		return nil, false
	}

	return ct, ok
}

func (t *tree) addChildTree(basepath string) {

	index := len(t.children)
	t.index[basepath] = index

	ct := &tree{children: []TreeEntry{}, index: map[string]int{}, base: basepath}
	t.children = append(t.children, ct)

}

func (t *tree) build() error {

	content := []byte{}

	for _, entry := range t.children {

		if err := entry.build(); err != nil {
			return err
		}

		e := []byte{}
		// the filemode
		e = append(e, []byte(entry.Permission())...)
		// a space
		e = append(e, 0x20)
		// the filename
		e = append(e, []byte(entry.Basename())...)
		// a null byte
		e = append(e, 0x00)
		// the oid packed into twenty bytes
		e = append(e, must(pack(entry.OID()))...)

		content = append(content, e...)

	}

	object, err := newObject(content, ClassTree)
	if err != nil {
		return err
	}

	t.object = object

	return nil

}

func (t *tree) Walk(f func(tree Object) error) error {

	for _, t := range t.children {
		if t, ok := t.(*tree); ok {
			t.Walk(f)
		}
	}

	return f(t)
}

func (t *tree) Basename() string {
	return t.base
}

func (t *tree) fullpath() string {
	return t.full
}

func (t *tree) Permission() Permission {
	return Directory
}

func (t *tree) OID() string {
	return t.id
}

func (t *tree) IsTree() bool {
	return true
}

func (t *tree) Children() []TreeEntry {
	return t.children
}

func pack(oid string) ([]byte, error) {

	packed := []byte{}

	for i := 0; i < len(oid); i += 2 {

		pair := oid[i : i+2]

		upper, err := strconv.ParseInt(string(pair[0]), 16, 8)
		if err != nil {
			return nil, err
		}

		lower, err := strconv.ParseInt(string(pair[1]), 16, 8)
		if err != nil {
			return nil, err
		}

		b := byte((upper << 4) + lower)

		packed = append(packed, b)
	}

	return packed, nil
}

func must[T any](v T, err error) T {

	if err != nil {
		panic(err)
	}

	return v
}
