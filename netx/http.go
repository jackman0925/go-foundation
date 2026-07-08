package netx

import (
	"net"
	"net/http"
)

// ClientIPFromHTTPRequest returns the host portion of request.RemoteAddr.
func ClientIPFromHTTPRequest(request *http.Request) string {
	if request == nil {
		return ""
	}

	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return request.RemoteAddr
}
