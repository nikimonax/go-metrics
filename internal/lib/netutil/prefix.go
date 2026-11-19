package netutil

import (
	"errors"
	"fmt"
	"net/netip"
)

func ParsePrefixes(addresses []string) ([]netip.Prefix, error) {
	var (
		prefixes []netip.Prefix
		errs     error
	)

	for _, entry := range addresses {
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			prefixes = append(prefixes, prefix)
			continue
		}

		if addr, err := netip.ParseAddr(entry); err == nil {
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}

		err := fmt.Errorf("invalid IP or CIDR: %s", entry)
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return nil, errs
	}

	return prefixes, nil
}
