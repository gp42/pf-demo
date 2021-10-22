package api_test

import (
	"net/http"
	"testing"

	"github.com/gp42/pf-demo/pkg/api"
)

// Return error message or empty string if error is nil
func errOrNil(err error) string {
	if err != nil {
		return err.Error()
	} else {
		return "<nil>"
	}
}

func TestRequestToIP(t *testing.T) {
	// Good IP4 address with port
	ip, err := api.RequestToIP(&http.Request{
		RemoteAddr: "10.0.0.1:50000",
	})
	expectIP := "10.0.0.1"
	if err != nil || ip != expectIP {
		t.Fatalf(`RequestToIP() cannot parse good ip4 with port, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Good IP4 address without port
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "10.0.0.1",
	})
	expectIP = "10.0.0.1"
	if err != nil || ip != expectIP {
		t.Fatalf(`RequestToIP() cannot parse good ip4 with port, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Bad IP4 address with wrong notation
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "10.0.0.1::80",
	})
	expectIP = ""
	if err == nil || ip != "" {
		t.Fatalf(`RequestToIP() parse bad ip4 returned no errors, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Bad IP4 address out of range
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "10.0.0.256",
	})
	expectIP = ""
	if err == nil || ip != "" {
		t.Fatalf(`RequestToIP() parse bad out of range ip4 returned no errors, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Good IP4 address in header
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "10.0.0.1",
		Header: http.Header{
			"X-Forwarded-For": []string{"10.0.0.2,10.0.0.3"},
		},
	})
	expectIP = "10.0.0.2"
	if err != nil || ip != expectIP {
		t.Fatalf(`RequestToIP() cannot parse good ip4 in header, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Good IP6 address with port
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "[2001:0db8:85aa:0000:0000:8a2e:0370:1111]:80",
	})
	expectIP = "2001:db8:85aa::8a2e:370:1111"
	if err != nil || ip != expectIP {
		t.Fatalf(`RequestToIP() cannot parse good ip6 with port, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Good IP6 address without port
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "2001:0db8:85aa:0000:0000:8a2e:0370:1111",
	})
	expectIP = "2001:db8:85aa::8a2e:370:1111"
	if err != nil || ip != expectIP {
		t.Fatalf(`RequestToIP() cannot parse good ip6 with port, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Bad IP6 address with wrong notation
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "2001:0db8:85aa:0000:0000:8a2e:0370:1111::80",
	})
	expectIP = ""
	if err == nil || ip != "" {
		t.Fatalf(`RequestToIP() parse bad ip6 returned no errors, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Bad IP6 address out of range
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "2001:0db8:85aa:0000:0000:8a2e:0370:ffff:ffff",
	})
	expectIP = ""
	if err == nil || ip != "" {
		t.Fatalf(`RequestToIP() parse bad out of range ip6 returned no errors, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}

	// Good IP6 address in header
	ip, err = api.RequestToIP(&http.Request{
		RemoteAddr: "2001:0db8:85aa:0000:0000:8a2e:0370:1110",
		Header: http.Header{
			"X-Forwarded-For": []string{"2001:0db8:85aa:0000:0000:8a2e:0370:1111,2001:0db8:85aa:0000:0000:8a2e:0370:1112"},
		},
	})
	expectIP = "2001:db8:85aa::8a2e:370:1111"
	if err != nil || ip != expectIP {
		t.Fatalf(`RequestToIP() cannot parse good ip4 in header, got: "%s", want "%s". error: %s`, ip, expectIP, errOrNil(err))
	}
}
