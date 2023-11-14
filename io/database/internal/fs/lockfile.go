package fs

import (
	"os"
)

type lockfile struct {
	filepath string
	lockfile *os.File
}

func NewLockfile(filepath string) (*lockfile, error) {

	f, err := os.OpenFile(filepath+".lock", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755)
	if err != nil {
		return nil, err
	}

	return &lockfile{filepath: filepath, lockfile: f}, nil
}

func (l *lockfile) Write(data []byte) error {

	if _, err := l.lockfile.Write(data); err != nil {
		return err
	}

	return nil
}

func (l *lockfile) Commit() error {

	defer l.lockfile.Close()

	if err := os.Rename(l.lockfile.Name(), l.filepath); err != nil {
		os.Remove(l.lockfile.Name())
		return err
	}

	return nil
}

func (l *lockfile) Release() error {
	return os.Remove(l.lockfile.Name())
}
