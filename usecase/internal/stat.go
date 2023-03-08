package internal

import (
	"errors"
	"os"
	"syscall"
)

func FileStat(path string) (*syscall.Stat_t, error) {

	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	statt, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, errors.New("cannot cast stat.Sys() to *syscall.Stat_t")
	}

	return statt, nil

}

func FileModePerm(filemode os.FileMode) os.FileMode {
	return filemode & 0777
}
