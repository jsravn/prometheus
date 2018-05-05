package main

import "github.com/gopherjs/gopherjs/js"

func uint8ArrayToBytes(buf uintptr) []byte {
	array := js.InternalObject(buf)
	slice := make([]byte, array.Length()) // TODO is this also off-by-one? humm
	js.InternalObject(slice).Set("$array", array)
	return slice
}

func uint8ArrayToString(buf uintptr) string {
	array := js.InternalObject(buf)
	length := array.Length()
	null := js.InternalObject(0)
	if length > 0 && array.Index(length-1) == null {
		length-- // TODO: understand from where/ why this nul terminator comes
	}
	slice := make([]byte, length)
	js.InternalObject(slice).Set("$array", array)
	return string(slice)
}
