package model

import (
	"io/fs"
	"syscall"
)

type File struct {
	Name       string
	Data       []byte
	Permission fs.FileMode
	Stat       *FileStat
}

func (f *File) IsExecutable() bool {
	return (f.Permission & 0111) == 0111
}

type FileStat struct {
	ctime, ctime_nsec, mtime, mtime_nsec uint32
	dev                                  uint32
	ino                                  uint32
	mode                                 uint32
	uid, gid                             uint32
	size                                 uint32
}

func NewFileStat(stat *syscall.Stat_t) *FileStat {

	return &FileStat{
		ctime:      uint32(stat.Ctimespec.Sec),
		ctime_nsec: uint32(stat.Ctimespec.Nsec),
		mtime:      uint32(stat.Mtimespec.Sec),
		mtime_nsec: uint32(stat.Mtimespec.Nsec),
		dev:        uint32(stat.Dev),
		ino:        uint32(stat.Ino),
		mode:       uint32(stat.Mode),
		uid:        uint32(stat.Uid),
		gid:        uint32(stat.Gid),
		size:       uint32(stat.Size),
	}

}
