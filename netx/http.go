package netx

import (
	"context"
	"errors"
	"net"
	"net/http"
)

var (
	// ErrHTTPServerRequired 表示 HTTP 服务实例为空。
	ErrHTTPServerRequired = errors.New("http server is required")
	// ErrShutdownContextRequired 表示关闭上下文为空。
	ErrShutdownContextRequired = errors.New("shutdown context is required")
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

// ShutdownHTTPServer 优雅关闭 HTTP 服务；超过 ctx 时限后强制关闭活动连接。
//
// 它不管理被 Hijack 的连接，例如 WebSocket；调用方应自行通知并等待这类连接退出。
func ShutdownHTTPServer(ctx context.Context, server *http.Server) error {
	if server == nil {
		return ErrHTTPServerRequired
	}
	if ctx == nil {
		return ErrShutdownContextRequired
	}

	if err := server.Shutdown(ctx); err != nil {
		// Shutdown 超时后，Close 确保活动连接不会阻止进程最终退出。
		if closeErr := server.Close(); closeErr != nil && !errors.Is(closeErr, http.ErrServerClosed) {
			return errors.Join(err, closeErr)
		}
		return err
	}
	return nil
}
