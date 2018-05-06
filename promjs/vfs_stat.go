package main

import (
	"syscall"
	"unsafe"

	"github.com/gopherjs/gopherjs/js"
)

func (vfs *virtualFileSystem) Stat(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	pathname := uint8ArrayToString(a1)
	buf := js.InternalObject(a2)
	statT := syscall.Stat_t{}

	file, exists := vfs.files[pathname]
	if !exists {
		return 0, 0, syscall.ENOENT
	}

	if file.isDir == 1 {
		statT.Mode = syscall.S_IFDIR | syscall.S_IRUSR | syscall.S_IWUSR | syscall.S_IXUSR
	} else {
		statT.Mode = syscall.S_IFREG | syscall.S_IRUSR | syscall.S_IWUSR
	}

	// only Mode is returned, to implement isDir()

	offset := int(unsafe.Offsetof(statT.Mode))
	hi := uint8(statT.Mode >> 8)
	lo := uint8(statT.Mode & 0xFF)

	buf.SetIndex(offset, lo)
	buf.SetIndex(offset+1, hi)

	return 0, 0, 0
}
