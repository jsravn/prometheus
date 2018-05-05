package main

import "syscall"

func (vfs *virtualFileSystem) Unlink(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	pathname := uint8ArrayToString(a1)

	_, exists := vfs.files[pathname]
	if !exists {
		return 0, 0, syscall.ENOENT
	}

	delete(vfs.files, pathname)

	return 0, 0, 0
}
