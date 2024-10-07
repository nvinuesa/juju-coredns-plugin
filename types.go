package juju

import (
	"fmt"

	"github.com/miekg/dns"
)

type JujuFQDN struct {
	Controller  string
	Model       string
	Application string
	Unit        string
}

func ParseJujuFQDN(qname string) (JujuFQDN, error) {
	labels, ok := dns.IsDomainName(qname)
	// FQDN will include juju.local., which is 6 labels total.
	if !ok || labels != 6 {
		return JujuFQDN{}, fmt.Errorf("invalid juju FQDN %q, labels %+v", qname, labels)
	}
	parts := dns.SplitDomainName(qname)
	if parts[4] != "juju" || parts[5] != "local" || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return JujuFQDN{}, fmt.Errorf("invalid juju FQDN %q: labels %+v cannot be empty, ", parts, qname)
	}
	return JujuFQDN{
		Controller:  parts[3],
		Model:       parts[2],
		Application: parts[1],
		Unit:        parts[0],
	}, nil
}

func (j *JujuFQDN) IsValid() bool {
	return j.Controller != "" && j.Model != "" && j.Application != "" && j.Unit != ""
}
