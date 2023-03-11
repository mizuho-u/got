package usecase

import (
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/database"
)

func InitDir(ctx GotContext) error {

	if err := os.MkdirAll(filepath.Join(ctx.GotRoot(), "objects"), os.ModeDir|0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(ctx.GotRoot(), "refs"), os.ModeDir|0755); err != nil {
		return err
	}

	return database.NewRefs(ctx.GotRoot()).UpdateHEAD("ref: refs/heads/main")

}
