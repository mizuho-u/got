package usecase

import (
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/database"
)

type ExitCode int

func InitDir(ctx GotContext) ExitCode {

	if err := os.MkdirAll(filepath.Join(ctx.GotRoot(), "objects"), os.ModeDir|0755); err != nil {
		return 128
	}

	if err := os.MkdirAll(filepath.Join(ctx.GotRoot(), "refs"), os.ModeDir|0755); err != nil {
		return 128
	}

	if err := database.NewRefs(ctx.GotRoot()).UpdateHEAD("ref: refs/heads/main"); err != nil {
		return 128
	}

	ctx.Out("Initialized empty Jit repository in " + ctx.WorkspaceRoot())

	return 0

}
