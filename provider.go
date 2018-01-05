package reverseoperator

import (
	secop "github.com/fardog/secureoperator"
	"github.com/miekg/dns"
)

// Provider is an interface representing a servicer of DNS queries.
type Provider interface {
	Query(*dns.Msg) (*secop.DNSResponse, error)
}
