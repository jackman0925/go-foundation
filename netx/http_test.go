package netx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
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

func TestShutdownHTTPServerGracefully(t *testing.T) {
	server, listener, serveErr := startHTTPServer(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := ShutdownHTTPServer(ctx, server); err != nil {
		t.Fatalf("ShutdownHTTPServer() error = %v", err)
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("Serve() error = %v, want %v", err, http.ErrServerClosed)
	}
}

func TestShutdownHTTPServerClosesActiveConnectionAfterTimeout(t *testing.T) {
	handlerStarted := make(chan struct{})
	handlerStopped := make(chan struct{})
	server, listener, serveErr := startHTTPServer(t, http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		close(handlerStarted)
		<-request.Context().Done()
		close(handlerStopped)
	}))
	defer listener.Close()

	connection, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer connection.Close()
	if _, err := connection.Write([]byte("GET / HTTP/1.1\r\nHost: example.test\r\n\r\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	select {
	case <-handlerStarted:
	case <-time.After(time.Second):
		t.Fatal("request handler did not start")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	err = ShutdownHTTPServer(ctx, server)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ShutdownHTTPServer() error = %v, want %v", err, context.DeadlineExceeded)
	}
	select {
	case <-handlerStopped:
	case <-time.After(time.Second):
		t.Fatal("active request was not interrupted after shutdown timeout")
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("Serve() error = %v, want %v", err, http.ErrServerClosed)
	}
}

func TestShutdownHTTPServerRejectsNilArguments(t *testing.T) {
	if err := ShutdownHTTPServer(context.Background(), nil); !errors.Is(err, ErrHTTPServerRequired) {
		t.Fatalf("ShutdownHTTPServer() error = %v, want %v", err, ErrHTTPServerRequired)
	}
	if err := ShutdownHTTPServer(nil, &http.Server{}); !errors.Is(err, ErrShutdownContextRequired) {
		t.Fatalf("ShutdownHTTPServer() error = %v, want %v", err, ErrShutdownContextRequired)
	}
}

func startHTTPServer(t *testing.T, handler http.Handler) (*http.Server, net.Listener, <-chan error) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	server := &http.Server{Handler: handler}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()
	return server, listener, serveErr
}
