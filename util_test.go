package reverseoperator

import (
	"net/url"
	"testing"
)

func TestURLToDNSQuestionValid(t *testing.T) {
	name := "example.com"
	typ := "2"
	u := url.URL{}
	v := u.Query()
	v.Set("name", name)
	v.Set("type", typ)
	u.RawQuery = v.Encode()

	q, err := urlToDNSQuestion(&u)
	if err != nil {
		t.Fatal(err)
	}

	if q.Name != name {
		t.Errorf("unexpected name %v", q.Name)
	}
	if q.Type != 2 {
		t.Errorf("unexpected type %v", q.Type)
	}
}

func TestURLToDNSQuestionBadName(t *testing.T) {
	name := ""
	typ := "1"
	u := url.URL{}
	v := u.Query()
	v.Set("name", name)
	v.Set("type", typ)
	u.RawQuery = v.Encode()

	_, err := urlToDNSQuestion(&u)
	if err != errNameInvalid {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestURLToDNSQuestionBadNameFragment(t *testing.T) {
	name := "wut..example.com"
	typ := "1"
	u := url.URL{}
	v := u.Query()
	v.Set("name", name)
	v.Set("type", typ)
	u.RawQuery = v.Encode()

	_, err := urlToDNSQuestion(&u)
	if err != errNameFragmentInvalid {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestURLToDNSQuestionBadType(t *testing.T) {
	name := "example.com"
	typ := "wut"
	u := url.URL{}
	v := u.Query()
	v.Set("name", name)
	v.Set("type", typ)
	u.RawQuery = v.Encode()

	_, err := urlToDNSQuestion(&u)
	if err != errTypeInvalid {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestURLToDNSQuestionTypeOutOfRange(t *testing.T) {
	name := "example.com"
	typ := "0"
	u := url.URL{}
	v := u.Query()
	v.Set("name", name)
	v.Set("type", typ)
	u.RawQuery = v.Encode()

	_, err := urlToDNSQuestion(&u)
	if err != errTypeOutOfRange {
		t.Errorf("unexpected error: %v", err)
	}
}
