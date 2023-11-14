package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mizuho-u/got/repository/object"
	"github.com/mizuho-u/got/types"
)

type Refs struct {
	gotpath string
}

func NewRefs(gotpath string) *Refs {
	return &Refs{gotpath}
}

func (r *Refs) heads(branch string) string {
	return filepath.Join(r.gotpath, "refs", "heads", branch)
}

func (r *Refs) Head() (object.Commit, error) {

	ref, err := r.resolveHead()
	if err != nil {
		return nil, err
	}

	return r.Ref(filepath.Base(ref))
}

const head string = `ref: (.+)`

func (r *Refs) resolveHead() (string, error) {

	f, err := os.Open(filepath.Join(r.gotpath, "HEAD"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	read, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	match := regexp.MustCompile(head).FindStringSubmatch(string(read))

	return match[1], nil
}

func (r *Refs) UpdateHeadRef(branchName types.BranchName) error {

	head, err := NewLockfile(filepath.Join(r.gotpath, "HEAD"))
	if err != nil {
		return err
	}

	ref := fmt.Sprintf("ref: refs/heads/%s", branchName.String())

	if err := head.Write([]byte(ref)); err != nil {
		return head.Release()
	}

	return head.Commit()

}

func (r *Refs) UpdateHeadCommit(commitId string) error {

	ref, err := r.resolveHead()
	if err != nil {
		return err
	}

	head, err := NewLockfile(r.heads(filepath.Base(ref)))
	if err != nil {
		return err
	}
	defer head.Commit()

	err = head.Write([]byte(commitId))
	if err != nil {
		return err
	}

	return nil
}

func (r *Refs) CreateBranch(branchName types.BranchName, oid string) error {

	if oid != "" {
		return r.UpdateRef(branchName.String(), oid)
	}

	head, err := r.Head()
	if err != nil {
		return err
	}

	return r.UpdateRef(branchName.String(), head.OID())
}

func (r *Refs) UpdateRef(name, oid string) error {

	path := r.heads(name)
	if isExist(path) {
		return fmt.Errorf("a branch named %s already exists", name)
	}

	ref, err := NewLockfile(path)
	if err != nil {
		return err
	}

	if err := ref.Write([]byte(oid)); err != nil {
		return ref.Release()
	}

	return ref.Commit()
}

func (r *Refs) Ref(branchName string) (object.Commit, error) {

	if branchName == "HEAD" {
		return r.Head()
	}

	f, err := os.Open(r.heads(branchName))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	commitId, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// mainが空なのはいいけど他はどうしようかなー
	if len(commitId) == 0 {
		return object.EmptyCommit(), nil
	}

	obj, err := load(r.gotpath, string(commitId))
	if err != nil {
		return nil, err
	}

	commit, err := object.ParseCommit(obj)
	if err != nil {
		return nil, err
	}

	return commit, nil
}
