package repository

import (
	"io/fs"

	"github.com/mizuho-u/got/repository/object"
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

func (s *FileStat) Permission() object.Permission {

	if fs.FileMode(s.mode).IsDir() {
		return object.Directory
	}

	if (s.mode & 0111) == 0111 {
		return object.ExecutableFile
	}

	return object.RegularFile
}

func NewFileStat(ctime, ctime_nsec, mtime, mtime_nsec, dev, ino, mode, uid, gid, size uint32) *FileStat {
	return &FileStat{
		ctime:      uint32(ctime),
		ctime_nsec: uint32(ctime_nsec),
		mtime:      uint32(mtime),
		mtime_nsec: uint32(mtime_nsec),
		dev:        uint32(dev),
		ino:        uint32(ino),
		mode:       uint32(mode),
		uid:        uint32(uid),
		gid:        uint32(gid),
		size:       uint32(size),
	}
}
