// +build js

package xxhash

import "github.com/gopherjs/gopherjs/js"

func Sum64(b []byte) uint64 {
	return h32(b)
}

func writeBlocks(x *xxh, b []byte) (res []byte) {
	return
}

func h64(b []byte) uint64 {
	hash := js.Global.Get("XXH").Call("h64", string(b), 0)
	return uint64(hash.Get("_a00").Uint64() +
		hash.Get("_a16").Uint64()<<16 +
		hash.Get("_a32").Uint64()<<32 +
		hash.Get("_a48").Uint64()<<48)
}

func h32(b []byte) uint64 {
	return uint64(js.Global.Get("XXH").Call("h64", string(b), 0).Int())
}
