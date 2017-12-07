package reverseoperator

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	secop "github.com/fardog/secureoperator"
)

func TestHandleGoodResponse(t *testing.T) {
	name := "example.com"
	ttl := uint32(100)
	data := "cool data"
	dnsresp := &secop.DNSResponse{
		Question: []secop.DNSQuestion{
			secop.DNSQuestion{Name: name, Type: 1}},
		Answer: []secop.DNSRR{
			secop.DNSRR{Name: "example.com", Type: 1, TTL: ttl, Data: data}},
	}
	provider := newFakeProvider(dnsresp, nil)
	h := NewHandler(provider, &HandlerOptions{})

	ts := httptest.NewServer(http.HandlerFunc(h.Handle))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("unable to parse url: %v", err)
	}
	q := u.Query()
	q.Set("name", name)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		t.Fatalf("unable to request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code %v", resp.StatusCode)
	}

	body := secop.GDNSResponse{}

	d := json.NewDecoder(resp.Body)
	err = d.Decode(&body)
	if err != nil {
		t.Fatalf("unable to parse response body: %v", err)
	}

	if l := len(body.Question); l != 1 {
		t.Fatalf("expected exactly one answer, got %v", l)
	}
	if n := body.Question[0].Name; n != "example.com" {
		t.Errorf("unexpected question name: %v", n)
	}
	if l := len(body.Answer); l != 1 {
		t.Fatalf("expected exactly one answer, got %v", l)
	}
	if n := body.Answer[0].Name; n != "example.com" {
		t.Errorf("unexpected answer name: %v", n)
	}
	if y := body.Answer[0].Type; y != 1 {
		t.Errorf("unexpected answer type: %v", y)
	}
	if y := body.Answer[0].TTL; y != ttl {
		t.Errorf("unexpected answer ttl: %v", y)
	}
	if d := body.Answer[0].Data; d != data {
		t.Errorf("unexpected answer data: %v", d)
	}
}

func TestHandleBadQuery(t *testing.T) {
	provider := newFakeProvider(nil, nil)
	h := NewHandler(provider, &HandlerOptions{})

	ts := httptest.NewServer(http.HandlerFunc(h.Handle))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unable to request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status code %v", resp.StatusCode)
	}
}

func TestHandleBadProvider(t *testing.T) {
	provider := newFakeProvider(nil, errors.New("frig"))
	h := NewHandler(provider, &HandlerOptions{})

	ts := httptest.NewServer(http.HandlerFunc(h.Handle))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("unable to parse url: %v", err)
	}
	q := u.Query()
	q.Set("name", "example.com")
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		t.Fatalf("unable to request: %v", err)
	}

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status code %v", resp.StatusCode)
	}
}

func newFakeProvider(resp *secop.DNSResponse, err error) *fakeProvider {
	return &fakeProvider{
		resp: resp,
		err:  err,
	}
}

type fakeProvider struct {
	req  *secop.DNSQuestion
	resp *secop.DNSResponse
	err  error
}

func (f *fakeProvider) Query(q secop.DNSQuestion) (*secop.DNSResponse, error) {
	f.req = &q
	return f.resp, f.err
}
