package netutil

import (
	"errors"
	"net"
)

var ErrExternalTarget = errors.New("unable to determine source IP for non local target")

func GetSourceIPFor(target net.IP) (net.IP, error) {
	if target.IsLoopback() {
		return target, nil
	}

	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.Contains(target) {
				return ipNet.IP, nil
			}
		}
	}

	return nil, ErrExternalTarget
}
