// Package autodiscovery implements JMAP service autodiscovery mechanism
// described in section 2.2 of JMAP Core RFC.
package autodiscovery

import (
	"errors"
	"strconv"

	"github.com/miekg/dns"
)

var ErrNoService = errors.New("autodiscovery: no DNS SRV record")

// Probe attempts to perform JMAP service autodiscovery by looking for DNS SRV
// entry. It returns URL of JMAP session resource. ErrNoService is returned in
// case there is no DNS SRV record for passed domain.
//
// Since Go standard DNS operations don't support SRV queries, custom DNS
// library is used and you have to provide related configuration options.
func Probe(cl *dns.Client, cfg *dns.ClientConfig, domain string) (string, error) {
	m := new(dns.Msg)
	m.SetQuestion("_jmap._tcp."+dns.Fqdn(domain), dns.TypeSRV)
	resp, _, err := cl.Exchange(m, cfg.Servers[0]+":"+cfg.Port)
	if err != nil {
		return "", err
	}

	for _, ans := range resp.Answer {
		srv, ok := ans.(*dns.SRV)
		if !ok {
			continue
		}

		return "https://" + srv.Target + ":" + strconv.Itoa(int(srv.Port)) + "/.well-known/jmap", nil
	}
	return "", ErrNoService
}
