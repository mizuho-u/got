package model

import (
	"io/fs"
	"syscall"

	"github.com/mizuho-u/got/model/internal"
	"github.com/mizuho-u/got/model/object"
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

	cspec := internal.Ctimespec(stat)
	mspec := internal.Mtimespec(stat)

	return &FileStat{
		ctime:      uint32(cspec.Sec),
		ctime_nsec: uint32(cspec.Nsec),
		mtime:      uint32(mspec.Sec),
		mtime_nsec: uint32(mspec.Nsec),
		dev:        uint32(stat.Dev),
		ino:        uint32(stat.Ino),
		mode:       uint32(stat.Mode),
		uid:        uint32(stat.Uid),
		gid:        uint32(stat.Gid),
		size:       uint32(stat.Size),
	}

}

func (s *FileStat) permission() object.Permission {

	if (s.mode & 0111) == 0111 {
		return object.ExecutableFile
	}

	return object.RegularFile

}
