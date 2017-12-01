package reverseoperator

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	secop "github.com/fardog/secureoperator"
)

var (
	errNameInvalid         = errors.New("length of name parameter must be between 1 and 253")
	errNameFragmentInvalid = errors.New("length of fragment in name parameter must be between 1 and 63")
	errTypeInvalid         = errors.New("type could not be parsed to an integer")
	errTypeOutOfRange      = errors.New("type was not within valid bounds 1 < x < 65535")
)

func urlToDNSQuestion(url *url.URL) (*secop.DNSQuestion, error) {
	v := url.Query()

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
	rtype, err := strconv.ParseUint(t, 10, 16)
	if t != "" && err != nil {
		return nil, errTypeInvalid
	} else {
		rtype = 1
	}
	if rtype < 1 || rtype > 65536 {
		return nil, errTypeOutOfRange
	}

	return &secop.DNSQuestion{
		Name: name,
		Type: uint16(rtype),
	}, nil
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
