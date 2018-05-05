package main

import (
	"syscall"

	"github.com/gopherjs/gopherjs/js"
)

// "Seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means
// relative to the current offset, and 2 means relative to the end. It returns the
// new offset and an error, if any.

func (vfs *virtualFileSystem) Seek(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	fd := a1
	offset := int(a2)
	whence := a3

	js.Global.Get("console").Call("debug", "::SEEK", fd, offset, whence)

	ref, ok := vfs.fds[fd]
	if !ok {
		return uintptr(minusOne), 0, syscall.EBADF
	}

	newPos := ref.pos
	switch whence {
	case 0:
		newPos = offset
	case 1:
		newPos += offset
	case 2:
		newPos = len(ref.file.data) + offset
	default:
		js.Global.Get("console").Call("warn", "SYS_LSEEK called with unexpected whence", whence)
		return uintptr(minusOne), 0, syscall.EINVAL
	}

	if newPos < 0 {
		return uintptr(minusOne), 0, syscall.EINVAL
	}
	ref.pos = newPos

	// grow the backing slice if we seek past the edge of the file? or let it panic if you do i/o out here
	if ref.pos > cap(ref.file.data) {
		js.Global.Get("console").Call("warn", "SYS_LSEEK past end of file", fd, ref.pos)
		// grow := make([]byte, ref.pos, ref.pos+1)
		// copy(grow, ref.file.data)
		// ref.file.data = grow
	}

	return uintptr(ref.pos), 0, 0
}
