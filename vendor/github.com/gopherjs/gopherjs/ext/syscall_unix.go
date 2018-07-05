// +build js,!windows

package ext

import "syscall"

func init() {
	syscall.RegisterCustomHandler(Syscall)
	syscall.RegisterCustomHandler6(Syscall6)
}
