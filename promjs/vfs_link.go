package main

import "syscall"

func (vfs *virtualFileSystem) Link(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	oldpath := uint8ArrayToString(a1)
	newpath := uint8ArrayToString(a2)

	file, exists := vfs.files[oldpath]
	if !exists {
		return 0, 0, syscall.ENOENT
	}
	if file.isDir == 1 {
		return 0, 0, syscall.EPERM
	}
	_, exists = vfs.files[newpath]
	if exists {
		return 0, 0, syscall.EEXIST
	}

	vfs.files[newpath] = file

	return 0, 0, 0
}
