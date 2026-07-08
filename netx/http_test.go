package netx

import (
	"net/http"
	"testing"
)

func TestClientIPFromHTTPRequest(t *testing.T) {
	req := &http.Request{RemoteAddr: "192.0.2.10:12345"}

	got := ClientIPFromHTTPRequest(req)
	if got != "192.0.2.10" {
		t.Fatalf("unexpected client IP: %q", got)
	}
}

func TestClientIPFromHTTPRequestHandlesRawIP(t *testing.T) {
	req := &http.Request{RemoteAddr: "192.0.2.10"}

	got := ClientIPFromHTTPRequest(req)
	if got != "192.0.2.10" {
		t.Fatalf("unexpected client IP: %q", got)
	}
}

func TestClientIPFromHTTPRequestHandlesNilRequest(t *testing.T) {
	if got := ClientIPFromHTTPRequest(nil); got != "" {
		t.Fatalf("expected empty IP, got %q", got)
	}
}

func TestClientIPFromHTTPRequestHandlesIPv6(t *testing.T) {
	req := &http.Request{RemoteAddr: "[2001:db8::1]:12345"}

	got := ClientIPFromHTTPRequest(req)
	if got != "2001:db8::1" {
		t.Fatalf("unexpected client IP: %q", got)
	}
}
