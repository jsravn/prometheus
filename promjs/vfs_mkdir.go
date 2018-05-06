package main

import (
	"os"
	"syscall"
)

func (vfs *virtualFileSystem) MkDir(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	pathname := uint8ArrayToString(a1)
	mode := os.FileMode(a2) // TODO?

	file, exists := vfs.files[pathname]
	if exists {
		return 0, 0, syscall.EEXIST
	}

	file = new(virtualFile)
	file.isDir = 1
	vfs.files[pathname] = file

	return 0, 0, 0
}

func (vfs *virtualFileSystem) initDir(paths ...string) {
	for _, pathname := range paths {
		file := new(virtualFile)
		file.isDir = 1
		vfs.files[pathname] = file
	}
}
