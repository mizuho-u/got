package usecase

import (
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/database"
)

func InitDir(path string) error {

	if err := os.MkdirAll(filepath.Join(path, ".git", "objects"), os.ModeDir|0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(path, ".git", "refs"), os.ModeDir|0755); err != nil {
		return err
	}

	return database.NewRefs(filepath.Join(path, ".git")).UpdateHEAD("ref: refs/heads/main")

}
