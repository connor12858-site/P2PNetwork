//go:build !android
// +build !android

package anet

import (
	"net"
)

// Non-Android build: simple wrappers around the standard net package.
func Interfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

func InterfaceAddrs() ([]net.Addr, error) {
	return net.InterfaceAddrs()
}

func InterfaceAddrsByInterface(ifi *net.Interface) ([]net.Addr, error) {
	if ifi == nil {
		return nil, nil
	}
	return ifi.Addrs()
}

func InterfaceByIndex(index int) (*net.Interface, error) {
	return net.InterfaceByIndex(index)
}

func InterfaceByName(name string) (*net.Interface, error) {
	return net.InterfaceByName(name)
}

func SetAndroidVersion(version uint) {}
