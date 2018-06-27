// +build js

package xxhash

import "github.com/gopherjs/gopherjs/js"

func Sum64(b []byte) uint64 {
	return h32(b)
}

func writeBlocks(x *xxh, b []byte) (res []byte) {
	return
}

// using the js xxh lib gives 2x speedup over the go version
func h64(b []byte) uint64 {
	hash := js.Global.Get("XXH").Call("h64", string(b), 0)
	return uint64(hash.Get("_a00").Uint64() +
		hash.Get("_a16").Uint64()<<16 +
		hash.Get("_a32").Uint64()<<32 +
		hash.Get("_a48").Uint64()<<48)
}

// this is 10% faster than h64 running "demo(4)"
func h32(b []byte) uint64 {
	lower := uint32(js.Global.Get("XXH").Call("h32", string(b), 0).Int())
	return uint64(lower) + uint64(lower)<<32
}
