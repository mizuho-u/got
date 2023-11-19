//go:build darwin
// +build darwin

package stat

import "syscall"

func Ctimespec(st *syscall.Stat_t) syscall.Timespec {
	return st.Ctimespec
}

func Mtimespec(st *syscall.Stat_t) syscall.Timespec {
	return st.Mtimespec
}
