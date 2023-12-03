package repository

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/mizuho-u/got/repository/object"
)

func (repo *repository) Diff(staged bool) ([]diff, error) {

	if staged {
		return repo.diffStaged()
	}

	diffs := []diff{}

	files, types := repo.WorkspaceChanges()

	for _, path := range files {

		var d diff

		switch types[path] {
		case statusFileDeleted:

			deleted := &diffDeleted{}

			deleted.AOID = repo.index.entries[path].oid
			deleted.AMode = string(repo.index.entries[path].permission())
			deleted.APath = filepath.Join("a", path)
			obj, err := repo.object.Load(deleted.AOID)
			if err != nil {
				return nil, err
			}
			deleted.AData = obj.Data()

			deleted.BOID = nullOID
			deleted.BPath = filepath.Join("b", path)
			deleted.BData = []byte(nullContents)

			d = deleted

		case statusFileModified:

			modified := &diffModified{}

			modified.AOID = repo.index.entries[path].oid
			modified.AMode = string(repo.index.entries[path].permission())
			modified.APath = path

			obj, err := repo.object.Load(modified.AOID)
			if err != nil {
				return nil, err
			}
			modified.AData = obj.Data()

			f := repo.workspace[path]
			f.Seek(0, io.SeekStart)
			data, err := io.ReadAll(f)
			if err != nil {
				return diffs, err
			}

			blob, err := object.NewBlob(path, data)
			if err != nil {
				return diffs, err
			}

			modified.BOID = blob.OID()
			modified.BMode = string(f.Stats().Permission())
			modified.BPath = path
			modified.BData = data

			d = modified

		}

		diffs = append(diffs, d)

	}

	return diffs, nil
}

func (repo *repository) diffStaged() ([]diff, error) {

	diffs := []diff{}

	files, types := repo.IndexChanges()

	for _, path := range files {

		var d diff

		switch types[path] {
		case statusFileModified:

			modified := &diffModified{}

			f := repo.head[path]

			modified.AOID = f.OID()
			modified.AMode = string(f.Permission())
			modified.APath = path
			adata, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}
			modified.AData = adata

			modified.BOID = repo.index.entries[path].oid
			modified.BMode = string(repo.index.entries[path].permission())
			modified.BPath = path

			obj, err := repo.object.Load(modified.BOID)
			if err != nil {
				return nil, err
			}
			modified.BData = obj.Data()

			d = modified

		case statusFileDeleted:

			deleted := &diffDeleted{}

			deleted.AOID = repo.head[path].OID()
			deleted.AMode = string(repo.head[path].Permission())
			deleted.APath = filepath.Join("a", path)

			obj, err := repo.object.Load(deleted.AOID)
			if err != nil {
				return nil, err
			}
			deleted.AData = obj.Data()

			deleted.BOID = nullOID
			deleted.BPath = filepath.Join("b", path)
			deleted.BData = []byte(nullContents)

			d = deleted

		case statusIndexAdded:

			added := &diffAdded{}

			added.AOID = nullOID
			added.APath = filepath.Join("a", path)
			added.AData = []byte(nullContents)

			added.BOID = repo.index.entries[path].oid
			added.BMode = string(repo.index.entries[path].permission())
			added.BPath = filepath.Join("b", path)
			obj, err := repo.object.Load(added.BOID)
			if err != nil {
				return nil, err
			}
			added.BData = obj.Data()

			d = added

		}

		diffs = append(diffs, d)

	}

	return diffs, nil
}

const (
	statusNone          status = " "
	statusIndexAdded    status = "A"
	statusFileDeleted   status = "D"
	statusFileModified  status = "M"
	statusFileUntracked status = ""
	statusUnchanged     status = statusNone + statusNone
)

type status string

func (s status) ShortFormat() string {
	return string(s)
}

func (s status) LongFormat() string {
	switch s {
	case statusIndexAdded:
		return "new file"
	case statusFileDeleted:
		return "deleted"
	case statusFileModified:
		return "modified"
	default:
		return ""
	}
}

const (
	nullPath     string = "/dev/null"
	nullOID      string = "0000000000000000000000000000000000000000"
	nullContents string = ""
)

type diff interface {
	PathLine() string
	ModeLine() string
	IndexLine() string
	FileLine() string
	Hunks() []*hunk
}

type diffModified struct {
	AOID  string
	AMode string
	APath string
	AData []byte
	BOID  string
	BMode string
	BPath string
	BData []byte
}

func (diff *diffModified) PathLine() string {
	return fmt.Sprintf("diff --git %s %s\n", filepath.Join("a", diff.APath), filepath.Join("b", diff.BPath))
}

func (diff *diffModified) ModeLine() string {

	l := ""

	if diff.AMode != diff.BMode {
		l += fmt.Sprintf("old mode %s\n", diff.AMode)
		l += fmt.Sprintf("new mode %s\n", diff.BMode)
	}

	return l
}

func (diff *diffModified) IndexLine() string {

	if diff.AOID == diff.BOID {
		return ""
	}

	if diff.AMode != diff.BMode {
		return fmt.Sprintf("index %s..%s\n", object.ShortOID(diff.AOID), object.ShortOID(diff.BOID))
	}

	return fmt.Sprintf("index %s..%s %s\n", object.ShortOID(diff.AOID), object.ShortOID(diff.BOID), diff.AMode)
}

func (diff *diffModified) FileLine() string {

	if diff.AOID == diff.BOID {
		return ""
	}

	l := fmt.Sprintf("--- %s\n", filepath.Join("a", diff.APath))
	l += fmt.Sprintf("+++ %s\n", filepath.Join("b", diff.BPath))

	return l

}

func (diff *diffModified) Hunks() []*hunk {

	al, _ := lines(bytes.NewBuffer(diff.AData))
	bl, _ := lines(bytes.NewBuffer(diff.BData))

	m := newMyers(al, bl)

	return m.diff().hunks()

}

type diffDeleted struct {
	AOID  string
	AMode string
	APath string
	AData []byte
	BOID  string
	BMode string
	BPath string
	BData []byte
}

func (diff *diffDeleted) PathLine() string {
	return fmt.Sprintf("diff --git %s %s\n", diff.APath, diff.BPath)
}

func (diff *diffDeleted) ModeLine() string {
	return fmt.Sprintf("deleted file mode %s\n", diff.AMode)
}

func (diff *diffDeleted) IndexLine() string {
	return fmt.Sprintf("index %s..%s\n", object.ShortOID(diff.AOID), object.ShortOID(diff.BOID))
}

func (diff *diffDeleted) FileLine() string {

	l := fmt.Sprintf("--- %s\n", diff.APath)
	l += fmt.Sprintf("+++ %s\n", nullPath)

	return l

}

func (diff *diffDeleted) Hunks() []*hunk {

	al, _ := lines(bytes.NewBuffer(diff.AData))
	bl, _ := lines(bytes.NewBuffer(diff.BData))

	m := newMyers(al, bl)

	return m.diff().hunks()

}

type diffAdded struct {
	AOID  string
	AMode string
	APath string
	AData []byte
	BOID  string
	BMode string
	BPath string
	BData []byte
}

func (diff *diffAdded) PathLine() string {
	return fmt.Sprintf("diff --git %s %s\n", diff.APath, diff.BPath)
}

func (diff *diffAdded) ModeLine() string {
	return fmt.Sprintf("new file mode %s\n", diff.BMode)
}

func (diff *diffAdded) IndexLine() string {
	return fmt.Sprintf("index %s..%s\n", object.ShortOID(diff.AOID), object.ShortOID(diff.BOID))
}

func (diff *diffAdded) FileLine() string {

	l := fmt.Sprintf("--- %s\n", nullPath)
	l += fmt.Sprintf("+++ %s\n", diff.BPath)

	return l
}

func (diff *diffAdded) Hunks() []*hunk {

	al, _ := lines(bytes.NewBuffer(diff.AData))
	bl, _ := lines(bytes.NewBuffer(diff.BData))

	m := newMyers(al, bl)

	return m.diff().hunks()

}
