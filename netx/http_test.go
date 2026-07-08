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
