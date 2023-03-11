//go:build linux
// +build linux

package internal

import "syscall"

func Ctimespec(st *syscall.Stat_t) syscall.Timespec {
	return st.Atim
}

func Mtimespec(st *syscall.Stat_t) syscall.Timespec {
	return st.Mtim
}
