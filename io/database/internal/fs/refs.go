package fs

import (
	"io"
	"os"
	"path/filepath"
)

type Refs struct {
	gotpath string
}

func NewRefs(gotpath string) *Refs {
	return &Refs{gotpath}
}

func (r *Refs) HEAD() (string, error) {

	f, err := os.Open(filepath.Join(r.gotpath, "HEAD"))
	if err == os.ErrNotExist {
		return "", nil
	}
	defer f.Close()

	read, err := io.ReadAll(f)
	if err == os.ErrNotExist {
		return "", err
	}

	// テスト時にgitコマンドがレポジトリを認識するようgot initで仮のHEADファイルを生成しているが、
	// refの実装はまだ追いついていないので実装するまで空で返す
	head := string(read)
	if head == "ref: refs/heads/main" {
		head = ""
	}

	return head, nil
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
