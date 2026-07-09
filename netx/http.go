package netx

import (
	"net"
	"net/http"
)

// ClientIPFromHTTPRequest 返回 request.RemoteAddr 中的 host 部分。
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
