package main

import (
	"syscall"

	"github.com/gopherjs/gopherjs/js"
)

func (vfs *virtualFileSystem) Fcntl(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	fd := a1
	cmd := a2
	arg := a3

	fileRef, exists := vfs.fds[fd]
	if !exists {
		return uintptr(minusOne), 0, syscall.EBADF
	}

	switch cmd {
	case syscall.F_DUPFD:
		if vfs.nextFD < arg {
			vfs.nextFD = arg
		}
		n := vfs.nextFD
		vfs.fds[n] = fileRef
		vfs.nextFD++
		return n, 0, 0
	case syscall.F_SETFL:
		// pretend to accept whatever flags were set
		return 0, 0, 0
	case syscall.F_GETFL:
		// pretend all the flags are O_NONBLOCK
		return syscall.O_NONBLOCK, 0, 0
	case syscall.F_PREALLOCATE:
		upper := js.InternalObject(a3).Get("Length").Get("$high").Int()
		if upper != 0 {
			return uintptr(minusOne), 0, syscall.EOVERFLOW
		}
		size := js.InternalObject(a3).Get("Length").Get("$low").Int()
		if size > fileRef.pos {
			js.Global.Get("console").Call("debug", "SYS_FCNTL", fd, "preallocating", size)
			grow := make([]byte, size, size+1)
			copy(grow, fileRef.file.data)
			fileRef.file.data = grow
		}
		return 0, 0, 0
	case syscall.F_FULLFSYNC:
		return 0, 0, 0
	}

	js.Global.Get("console").Call("warn", "SYS_FCNTL unhandled cmd", cmd)

	return 0, 0, 0
}
