package main

import (
	"syscall"

	"github.com/gopherjs/gopherjs/js"
)

// Write a file: http://man7.org/linux/man-pages/man2/write.2.html
//
//       ssize_t write(int fd, const void *buf, size_t count);
//
func (vfs *virtualFileSystem) Write(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	fd := a1
	buf := uint8ArrayToBytes(a2)
	cnt := a3

	// write to stdout/stdin
	switch fd {
	case uintptr(syscall.Stdout), uintptr(syscall.Stderr):
		js.Global.Get("console").Call("log", string(buf))
		return cnt, 0, 0
	}

	// find our file descriptor
	ref, ok := vfs.fds[fd]
	if !ok {
		return uintptr(minusOne), 0, syscall.EBADF
	}

	// append to the file data and move the cursor
	ref.file.data = append(ref.file.data[:ref.pos], buf...)
	ref.pos += len(buf)

	return cnt, 0, 0
}
