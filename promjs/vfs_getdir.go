package main

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/gopherjs/gopherjs/js"
)

/*
	man 2 getdirentries

	Takes a file descriptor of a directory and incrementally copies Dirent pages for contents of directory.
	Dirent pages are run time length encoded.
	CAVEAT: Our implementation does not do continuation, must fit within first buffer.
	Returns the number of bytes filled.
	Caller only stops when return is <= 0
*/

func (vfs *virtualFileSystem) Getdir(a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	fd := a1
	dest := js.InternalObject(a2)
	limit := int(a3) // 4096
	// a4 is by spec an optional pointer to fd.pos, but in Go this is reset between calls so useless

	js.Global.Get("console").Call("debug", "::GETDIRENTRIES64", fd, dest, limit)

	// print the file system:
	// for path := range vfs.files {
	// 	js.Global.Get("console").Call("debug", "  ", path)
	// }

	// remove
	vfs.BREAKS++
	if vfs.BREAKS >= 20 {
		js.Debugger()
	}

	ref, ok := vfs.fds[fd]
	if !ok || ref.file.isDir != 1 {
		return 0, 0, syscall.EBADF
	}

	if ref.pos != 0 {
		// we only have the previous page to report
		return 0, 0, 0
	}

	// find the directory canonical name, and match things in there
	prefix := ""
	for path, f := range vfs.files {
		if f == ref.file {
			prefix = path + "/"
			js.Global.Get("console").Call("debug", "getdir looking for", prefix)
			break
		}
	}

	length := 0
	inode := uint64(0)
	for path := range vfs.files {
		inode++
		if strings.HasPrefix(path, prefix) {
			// found a file
			js.Global.Get("console").Call("debug", "getdir found", path)

			dirent := syscall.Dirent{}

			filename := filepath.Base(path)
			for n, c := range []byte(filename) {
				dirent.Name[n] = int8(c)
			}
			dirent.Namlen = uint16(len(filename))

			if vfs.files[path].isDir == 1 {
				dirent.Type = syscall.DT_DIR
			} else {
				dirent.Type = syscall.DT_REG
			}

			dirent.Ino = inode

			// spec mentions optional padding mod 4, so add 1 to 4 zeros at the end
			// it might be sufficient to just add one, but this is small
			recordLen := (int(unsafe.Offsetof(dirent.Name)) + int(dirent.Namlen) + 4) / 4 * 4
			dirent.Reclen = uint16(recordLen)

			if length+recordLen > limit {
				js.Global.Get("console").Call("warn", "getdir too many records", path)
				return 0, 0, syscall.EFAULT
			}

			// copy the bytes, there's probably a better way
			// but the js interop doesn't seem to play well with raw access
			src := bytes.Buffer{}
			binary.Write(&src, binary.LittleEndian, dirent)

			for n := 0; n < recordLen; n++ {
				c, err := src.ReadByte()
				if err != nil {
					break
				}
				dest.SetIndex(length+n, c)
			}
			length += recordLen

			// api requires fd.pos to point to next block
			ref.pos += length
		}
	}

	return uintptr(length), 0, 0
}
