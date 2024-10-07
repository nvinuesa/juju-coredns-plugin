package juju

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

func (j *Juju) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.QName()

	// Ignore non juju related queries.
	zone := plugin.Zones(zones).Matches(qname)
	if zone == "" {
		return plugin.NextOrFailure(j.Name(), j.Next, ctx, w, r)
	}

	m := &dns.Msg{}
	m.SetReply(r)

	jujuFQDN, err := ParseJujuFQDN(qname)
	if err != nil {
		log.Errorf("parsing juju FQDN %+v: %w", jujuFQDN, err)
		m.SetRcode(state.Req, dns.RcodeNameError)
		m.Authoritative = true
		w.WriteMsg(m)
		return dns.RcodeBadName, dns.ErrFqdn
	}

	address, err := j.GetAddress(ctx, jujuFQDN)
	if err != nil {
		log.Errorf("getting address for juju FQDN %+v: %w", jujuFQDN, err)
		m.SetRcode(state.Req, dns.RcodeNameError)
		m.Authoritative = true
		w.WriteMsg(m)
		return dns.RcodeBadName, err
	}

	switch state.QType() {
	case dns.TypeA:
		hdr := dns.RR_Header{Name: qname, Ttl: j.Ttl, Class: dns.ClassINET, Rrtype: dns.TypeA}
		m.Answer = []dns.RR{&dns.A{Hdr: hdr, A: address.To4()}}
		log.Debugf("Answering A query for juju FQDN %+v", m.Answer)
	case dns.TypeAAAA:
		// No IPv6 implemented.
	case dns.TypeCNAME:
	default:
	}

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}
