package reverseoperator

import (
	"strings"

	secop "github.com/fardog/secureoperator"
	"github.com/miekg/dns"
)

var exchange = dns.Exchange

func NewDNSProvider(servers secop.Endpoints) (*DNSProvider, error) {
	return &DNSProvider{
		servers: servers,
	}, nil
}

type DNSProvider struct {
	servers secop.Endpoints
}

func (c *DNSProvider) Query(msg *dns.Msg) (*secop.DNSResponse, error) {
	// we need to look it up
	server := c.servers.Random()

	r, err := exchange(msg, server.String())
	if err != nil {
		return nil, err
	}

	return &secop.DNSResponse{
		Truncated:          r.MsgHdr.Truncated,
		RecursionDesired:   r.MsgHdr.RecursionDesired,
		RecursionAvailable: r.MsgHdr.RecursionAvailable,
		AuthenticatedData:  r.MsgHdr.AuthenticatedData,
		CheckingDisabled:   r.MsgHdr.CheckingDisabled,
		ResponseCode:       r.MsgHdr.Rcode,
		Question:           questionToDNSQuestion(r.Question),
		Answer:             rrToDNSRR(r.Answer),
		Authority:          rrToDNSRR(r.Ns),
		Extra:              rrToDNSRR(r.Extra),
	}, nil
}

func questionToDNSQuestion(qs []dns.Question) []secop.DNSQuestion {
	var dqs []secop.DNSQuestion
	for _, q := range qs {
		dqs = append(dqs, secop.DNSQuestion{
			Name: q.Name,
			Type: q.Qtype,
		})
	}

	return dqs
}

func rrToDNSRR(rrs []dns.RR) []secop.DNSRR {
	var drs []secop.DNSRR
	for _, rr := range rrs {
		drs = append(drs, secop.DNSRR{
			Name: rr.Header().Name,
			Type: rr.Header().Rrtype,
			TTL:  rr.Header().Ttl,
			Data: strings.Replace(rr.String(), rr.Header().String(), "", 1),
		})
	}

	return drs
}
