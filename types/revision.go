package types

import (
	"fmt"
	"regexp"
	"strconv"
)

type Revision interface {
	String() string
	Resolve(resolver resolver) (ObjectID, error)
}

const (
	parentOperator   = `^(.+)\^$`
	ancestorOperator = `^(.+)~(\d+)$`
)

var refAlias = map[string]string{
	"@": "HEAD",
}

func NewRevision(s string) (Revision, error) {

	if s == "" {
		return &emptyRevision{}, nil
	}

	return parseRevision(s)
}

type emptyRevision struct{}

func (r *emptyRevision) String() string {
	return ""
}

func (r *emptyRevision) Resolve(resolver resolver) (ObjectID, error) {
	return "", nil
}

func parseRevision(s string) (Revision, error) {

	if match := regexp.MustCompile(parentOperator).FindStringSubmatch(s); len(match) != 0 {

		rev, err := parseRevision(match[1])
		if err != nil {
			return nil, err
		}

		return newParent(rev)

	} else if match := regexp.MustCompile(ancestorOperator).FindStringSubmatch(s); len(match) != 0 {

		rev, err := parseRevision(match[1])
		if err != nil {
			return nil, err
		}

		n, err := strconv.Atoi(match[2])
		if err != nil {
			return nil, err
		}

		return newAncestor(rev, n)

	} else if ref, err := newRef(s); err == nil {
		return ref, nil
	}

	return nil, fmt.Errorf("invalid revision %s", s)

}

type ref struct {
	name string
}

func newRef(s string) (*ref, error) {

	if err := validateRefName(s); err != nil {
		return nil, err
	}

	if v, ok := refAlias[s]; ok {
		s = v
	}

	return &ref{s}, nil
}

func (r *ref) String() string {
	return fmt.Sprintf("(ref %s)", r.name)
}

func (r *ref) Resolve(resolver resolver) (ObjectID, error) {
	return resolver.Ref(r.name)
}

type parent struct {
	rev Revision
}

func newParent(rev Revision) (*parent, error) {
	return &parent{rev}, nil
}

func (p *parent) String() string {
	return fmt.Sprintf("(parent %s)", p.rev)
}

func (p *parent) Resolve(resolver resolver) (ObjectID, error) {

	oid, err := p.rev.Resolve(resolver)
	if err != nil {
		return "", err
	}

	return resolver.Parent(oid)
}

type ancestor struct {
	rev Revision
	n   int
}

func newAncestor(rev Revision, n int) (*ancestor, error) {
	return &ancestor{rev, n}, nil
}

func (a *ancestor) String() string {
	return fmt.Sprintf("(ancestor %s %d)", a.rev, a.n)
}

func (a *ancestor) Resolve(resolver resolver) (ObjectID, error) {

	oid, err := a.rev.Resolve(resolver)
	if err != nil {
		return "", err
	}

	for i := 0; i < a.n; i++ {

		oid, err = resolver.Parent(oid)
		if err != nil {
			return "", err
		}

	}

	return oid, nil
}

type resolver interface {
	Ref(name string) (ObjectID, error)
	Parent(oid ObjectID) (ObjectID, error)
}
