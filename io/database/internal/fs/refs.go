package fs

import (
	"io"
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/repository/object"
)

type Refs struct {
	gotpath string
}

func NewRefs(gotpath string) *Refs {
	return &Refs{gotpath}
}

func (r *Refs) Head() (object.Commit, error) {

	f, err := os.Open(filepath.Join(r.gotpath, "HEAD"))
	if err == os.ErrNotExist {
		return nil, nil
	}
	defer f.Close()

	read, err := io.ReadAll(f)
	if err == os.ErrNotExist {
		return nil, err
	}

	// テスト時にgitコマンドがレポジトリを認識するようgot initで仮のHEADファイルを生成しているが、
	// refの実装はまだ追いついていないので実装するまで空で返す
	head := string(read)
	if head == "ref: refs/heads/main" {
		return object.EmptyCommit(), nil
	}

	obj, err := load(r.gotpath, head)
	if err != nil {
		return nil, err
	}

	commit, err := object.ParseCommit(obj)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (r *Refs) UpdateHEAD(commitId string) error {

	head, err := NewLockfile(filepath.Join(r.gotpath, "HEAD"))
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
