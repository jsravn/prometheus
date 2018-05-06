package main

import (
	"syscall"
)

// "Truncate changes the size of the file.
// It does not change the I/O offset."

func (vfs *virtualFileSystem) Truncate(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	fd := a1
	size := int(a2)

	ref, ok := vfs.fds[fd]
	if !ok {
		return uintptr(minusOne), 0, syscall.EBADF
	}

	if size < ref.pos {
		// shrink
		ref.file.data = ref.file.data[:size]
	} else if size > ref.pos {
		// grow

		// ... here I think we just quietly ignore the request

		// grow := make([]byte, size, size+1)
		// copy(grow, ref.file.data)
		// ref.file.data = grow
	}

	return 0, 0, 0
}
