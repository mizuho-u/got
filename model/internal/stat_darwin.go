//go:build darwin
// +build darwin

package internal

import "syscall"

func Ctimespec(st *syscall.Stat_t) syscall.Timespec {
	return st.Ctimespec
}

func Mtimespec(st *syscall.Stat_t) syscall.Timespec {
	return st.Mtimespec
}
