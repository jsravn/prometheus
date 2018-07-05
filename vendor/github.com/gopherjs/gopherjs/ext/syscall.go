package ext

import (
	"syscall"

	"github.com/gopherjs/gopherjs/js"
)

var minusOne = -1

type (
	// SyscallHandler is a handler callback for syscalls
	SyscallHandler = func(a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
	// SyscallHandler6 is a handler callback for the 6 argument variant of syscalls
	SyscallHandler6 = func(a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)
)

var (
	handler3s = map[uintptr]SyscallHandler{}
	handler6s = map[uintptr]SyscallHandler6{}
)

// Syscall handles a syscall
func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	handler, ok := handler3s[trap]
	if ok {
		return handler(a1, a2, a3)
	}

	js.Global.Get("console").Call("warn", "syscall not implemented", trap, a1, a2, a3)

	return uintptr(minusOne), 0, syscall.EACCES
}

// RegisterSyscallHandler registers a syscall handler for the given syscall
func RegisterSyscallHandler(trap uintptr, handler SyscallHandler) {
	handler3s[trap] = handler
}

// Syscall6 handles a syscall (the 6 argument variety)
func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	handler, ok := handler6s[trap]
	if ok {
		return handler(a1, a2, a3, a4, a5, a6)
	}

	js.Global.Get("console").Call("warn", "syscall6 not implemented", trap, a1, a2, a3)

	return uintptr(minusOne), 0, syscall.EACCES
}

// RegisterSyscallHandler6 registers a syscall handler for the given syscall
func RegisterSyscallHandler6(trap uintptr, handler SyscallHandler6) {
	handler6s[trap] = handler
}
