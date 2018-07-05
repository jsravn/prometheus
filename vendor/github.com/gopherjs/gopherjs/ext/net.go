// +build !js

package ext

import "net"

// RegisterListenFunc registers a listen function for net
func RegisterListenFunc(handler func(string, string) (net.Listener, error)) {
}
