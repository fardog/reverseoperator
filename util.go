package reverseoperator

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/miekg/dns"

	secop "github.com/fardog/secureoperator"
)

var (
	errNameInvalid         = errors.New("length of name parameter must be between 1 and 253")
	errNameFragmentInvalid = errors.New("length of fragment in name parameter must be between 1 and 63")
	errTypeInvalid         = errors.New("type could not be mapped to a valid DNS record type")
	errTypeOutOfRange      = errors.New("type was not within valid bounds 1 < x < 65535")
	errInvalidCIDR         = errors.New("invalid CIDR was provided")
	errBadRemoteAddress    = errors.New("bad remote address received")
	errBadIPAddress        = errors.New("bad ip address")
)

func reqToDNSMsg(r *http.Request) (*dns.Msg, error) {
	v := r.URL.Query()

	name := v.Get("name")
	if l := len(name); l < 1 || l > 253 {
		return nil, errNameInvalid
	}
	for _, f := range strings.Split(name, ".") {
		if l := len(f); l < 1 || l > 63 {
			return nil, errNameFragmentInvalid
		}
	}

	t := v.Get("type")
	var rtype uint16
	if t == "" {
		rtype = 1
	} else if st, ok := dns.StringToType[strings.ToUpper(t)]; ok {
		rtype = st
	} else if rt, err := strconv.ParseUint(t, 10, 16); err != nil {
		return nil, errTypeInvalid
	} else if rt < 1 {
		return nil, errTypeOutOfRange
	} else {
		rtype = uint16(rt)
	}

	msg := dns.Msg{}
	msg.SetQuestion(dns.Fqdn(name), rtype)

	if e := v.Get("edns_client_subnet"); e != secop.GoogleEDNSSentinelValue {
		const (
			IPv4Family = 1
			IPv6Family = 2
		)

		var network *net.IPNet
		var err error
		// if not provided, use the remote address
		if e == "" {
			h, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				return nil, errBadRemoteAddress
			}
			ip := net.ParseIP(h)
			if ip == nil {
				return nil, errBadRemoteAddress
			}

			var mask net.IPMask
			if ip.To4() != nil {
				mask = net.CIDRMask(24, 32)
			} else if ip.To16() != nil {
				mask = net.CIDRMask(36, 128)
			} else {
				return nil, errBadIPAddress
			}

			network = &net.IPNet{
				IP:   ip,
				Mask: mask,
			}
		} else {
			_, network, err = net.ParseCIDR(e)
			if err != nil {
				return nil, errInvalidCIDR
			}
		}

		// ok, we have what we need; construct the record
		o := new(dns.OPT)
		o.Hdr.Name = "."
		o.Hdr.Rrtype = dns.TypeOPT
		e := new(dns.EDNS0_SUBNET)
		e.Code = dns.EDNS0SUBNET
		if network.IP.To4() != nil {
			e.Family = IPv4Family
		} else if network.IP.To16() != nil {
			e.Family = IPv6Family
		} else {
			return nil, errBadIPAddress
		}
		msize, _ := network.Mask.Size()
		e.SourceNetmask = uint8(msize)
		e.SourceScope = 0
		e.Address = network.IP.Mask(network.Mask)
		o.Option = append(o.Option, e)

		msg.Extra = append(msg.Extra, o)
	}

	return &msg, nil
}

func fromDNStoGDNS(d *secop.DNSResponse) *secop.GDNSResponse {
	return &secop.GDNSResponse{
		Status:     int32(d.ResponseCode),
		TC:         d.Truncated,
		RD:         d.RecursionDesired,
		RA:         d.RecursionAvailable,
		AD:         d.AuthenticatedData,
		CD:         d.CheckingDisabled,
		Question:   fromDNSQuestionToGDNSQuestion(d.Question),
		Answer:     fromDNSRRsToGDNSRRs(d.Answer),
		Authority:  fromDNSRRsToGDNSRRs(d.Authority),
		Additional: fromDNSRRsToGDNSRRs(d.Extra),
	}
}

func fromDNSRRsToGDNSRRs(d []secop.DNSRR) secop.GDNSRRs {
	var g []secop.GDNSRR
	for _, r := range d {
		g = append(g, secop.GDNSRR(r))
	}
	return g
}

func fromDNSQuestionToGDNSQuestion(d []secop.DNSQuestion) secop.GDNSQuestions {
	var g []secop.GDNSQuestion
	for _, r := range d {
		g = append(g, secop.GDNSQuestion(r))
	}
	return g
}
