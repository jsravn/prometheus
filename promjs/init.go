package main

import (
	"syscall"

	"github.com/gopherjs/gopherjs/ext"
)

var vfs interface {
	Close(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
	Fcntl(a1, a2, a3 uintptr) (r1, r2 uintptr, error syscall.Errno)
	Getdir(a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)
	Link(a1, a2, a3 uintptr) (r1, r2 uintptr, error syscall.Errno)
	MkDir(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
	Open(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
	Read(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
	Seek(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
	Stat(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
	Truncate(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
	Unlink(a1, a2, a3 uintptr) (r1, r2 uintptr, error syscall.Errno)
	Write(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
} = newVirtualFileSystem()

func init() {
	ext.RegisterSyscallHandler(syscall.SYS_CLOSE, vfs.Close)
	ext.RegisterSyscallHandler(syscall.SYS_FCNTL, vfs.Fcntl)
	ext.RegisterSyscallHandler(syscall.SYS_FTRUNCATE, vfs.Truncate)
	ext.RegisterSyscallHandler(syscall.SYS_LINK, vfs.Link)
	ext.RegisterSyscallHandler(syscall.SYS_LSEEK, vfs.Seek)
	ext.RegisterSyscallHandler(syscall.SYS_LSTAT64, vfs.Stat)
	ext.RegisterSyscallHandler(syscall.SYS_MKDIR, vfs.MkDir)
	ext.RegisterSyscallHandler(syscall.SYS_OPEN, vfs.Open)
	ext.RegisterSyscallHandler(syscall.SYS_READ, vfs.Read)
	ext.RegisterSyscallHandler(syscall.SYS_STAT64, vfs.Stat)
	ext.RegisterSyscallHandler(syscall.SYS_UNLINK, vfs.Unlink)
	ext.RegisterSyscallHandler(syscall.SYS_WRITE, vfs.Write)
	ext.RegisterSyscallHandler6(syscall.SYS_GETDIRENTRIES64, vfs.Getdir)

	// pid is 2, ppid is 1
	ext.RegisterSyscallHandler(syscall.SYS_GETPID, func(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		return 2, 0, 0
	})
	ext.RegisterSyscallHandler(syscall.SYS_GETPPID, func(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		return 1, 0, 0
	})

	ext.RegisterSyscallHandler(syscall.SYS_FSYNC, func(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
		return 0, 0, 0
	})
}
