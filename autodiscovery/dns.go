// Package autodiscovery implements JMAP service autodiscovery mechanism
// described in section 2.2 of JMAP Core RFC.
package autodiscovery

import (
	"errors"
	"net"
	"strconv"
)

var ErrNoService = errors.New("autodiscovery: no DNS SRV record")

// Probe attempts to perform JMAP service autodiscovery by looking for DNS SRV
// entry. It returns URL of JMAP session resource. ErrNoService is returned in
// case there is no DNS SRV record for passed domain.
func Probe(domain string) (string, error) {
	_, addrs, err := net.LookupSRV("jmap", "tcp", domain)
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		return "https://" + addr.Target + ":" + strconv.Itoa(int(addr.Port)) + "/.well-known/jmap", nil
	}
	return "", ErrNoService
}
