package mailx

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

type smtpCapture struct {
	recipients []string
	data       string
}

func TestSMTPSenderSendsEnvelopeRecipientsAndMessage(t *testing.T) {
	addr, capture, stop := startFakeSMTPServer(t)
	defer stop()

	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort returned error: %v", err)
	}
	var port int
	if _, err := fmt.Sscanf(portText, "%d", &port); err != nil {
		t.Fatalf("parse port returned error: %v", err)
	}

	sender, err := NewSMTPSender(SMTPOptions{
		Host:       host,
		Port:       port,
		From:       "default@example.com",
		Encryption: EncryptionNone,
		Timeout:    time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	err = sender.Send(context.Background(), Message{
		To:      []string{"to@example.com"},
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"hidden@example.com"},
		Subject: "hello",
		Text:    "content",
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	got := <-capture
	wantRecipients := []string{"to@example.com", "cc@example.com", "hidden@example.com"}
	if len(got.recipients) != len(wantRecipients) {
		t.Fatalf("expected recipients %v, got %v", wantRecipients, got.recipients)
	}
	for i := range wantRecipients {
		if got.recipients[i] != wantRecipients[i] {
			t.Fatalf("recipient %d mismatch: want %s got %s", i, wantRecipients[i], got.recipients[i])
		}
	}
	if !strings.Contains(got.data, "From: <default@example.com>") {
		t.Fatalf("expected default From header, got:\n%s", got.data)
	}
	if strings.Contains(got.data, "hidden@example.com") {
		t.Fatalf("Bcc recipient leaked into message data:\n%s", got.data)
	}
}

func TestSMTPSenderRejectsOversizedMessage(t *testing.T) {
	sender, err := NewSMTPSender(SMTPOptions{
		Host:            "127.0.0.1",
		Port:            25,
		From:            "from@example.com",
		MaxMessageBytes: 10,
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	err = sender.Send(context.Background(), Message{
		To:      []string{"to@example.com"},
		Subject: "hello",
		Text:    "content larger than ten bytes",
	})
	if err == nil {
		t.Fatal("expected oversized message error")
	}
}

func TestSMTPSenderReturnsValidationErrorWhenDefaultFromMissing(t *testing.T) {
	sender, err := NewSMTPSender(SMTPOptions{
		Host: "127.0.0.1",
		Port: 25,
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	err = sender.Send(context.Background(), Message{
		To:      []string{"to@example.com"},
		Subject: "hello",
		Text:    "content",
	})
	if err == nil {
		t.Fatal("expected missing from validation error")
	}
}

func TestNewSMTPSenderRejectsInvalidOptions(t *testing.T) {
	if _, err := NewSMTPSender(SMTPOptions{Port: 25, From: "from@example.com"}); err == nil {
		t.Fatal("expected missing host error")
	}
	if _, err := NewSMTPSender(SMTPOptions{Host: "localhost", From: "from@example.com"}); err == nil {
		t.Fatal("expected missing port error")
	}
	if _, err := NewSMTPSender(SMTPOptions{Host: "localhost", Port: 25, From: "bad"}); err == nil {
		t.Fatal("expected invalid from error")
	}
	if _, err := NewSMTPSender(SMTPOptions{Host: "localhost", Port: 25, MaxMessageBytes: -1}); err == nil {
		t.Fatal("expected invalid max message bytes error")
	}
}

func TestSMTPSenderRejectsInvalidEncryption(t *testing.T) {
	sender, err := NewSMTPSender(SMTPOptions{
		Host:       "127.0.0.1",
		Port:       25,
		From:       "from@example.com",
		Encryption: Encryption(99),
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	if err := sender.Send(context.Background(), validMessage()); err == nil {
		t.Fatal("expected invalid encryption error")
	}
}

func TestSMTPSenderTLSConfig(t *testing.T) {
	sender, err := NewSMTPSender(SMTPOptions{
		Host: "smtp.example.com",
		Port: 465,
		From: "from@example.com",
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	cfg := sender.tlsConfig()
	if cfg.ServerName != "smtp.example.com" {
		t.Fatalf("expected server name smtp.example.com, got %s", cfg.ServerName)
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected TLS 1.2 min version, got %d", cfg.MinVersion)
	}

	custom := &tls.Config{ServerName: "custom.example.com", MinVersion: tls.VersionTLS13}
	sender.options.TLSConfig = custom
	if got := sender.tlsConfig(); got != custom {
		t.Fatal("expected custom TLS config")
	}
}

func TestSMTPSenderReturnsServerRejectErrors(t *testing.T) {
	tests := []string{"MAIL", "RCPT", "DATA", "ENDDATA", "QUIT"}

	for _, reject := range tests {
		t.Run(reject, func(t *testing.T) {
			addr, _, stop := startFakeSMTPServerWithReject(t, reject)
			defer stop()

			host, port := splitSMTPAddr(t, addr)

			sender, err := NewSMTPSender(SMTPOptions{
				Host:       host,
				Port:       port,
				From:       "from@example.com",
				Encryption: EncryptionNone,
				Timeout:    time.Second,
			})
			if err != nil {
				t.Fatalf("NewSMTPSender returned error: %v", err)
			}

			if err := sender.Send(context.Background(), validMessage()); err == nil {
				t.Fatal("expected server reject error")
			}
		})
	}
}

func TestSMTPSenderReturnsAuthError(t *testing.T) {
	addr, _, stop := startFakeSMTPServerWithReject(t, "AUTH")
	defer stop()

	host, port := splitSMTPAddr(t, addr)
	sender, err := NewSMTPSender(SMTPOptions{
		Host:       host,
		Port:       port,
		Username:   "user",
		Password:   "secret",
		From:       "from@example.com",
		Encryption: EncryptionNone,
		Timeout:    time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	if err := sender.Send(context.Background(), validMessage()); err == nil {
		t.Fatal("expected auth error")
	}
}

func TestSMTPSenderReturnsDialError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	host, port := splitSMTPAddr(t, addr)
	sender, err := NewSMTPSender(SMTPOptions{
		Host:       host,
		Port:       port,
		From:       "from@example.com",
		Encryption: EncryptionNone,
		Timeout:    50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	if err := sender.Send(context.Background(), validMessage()); err == nil {
		t.Fatal("expected dial error")
	}
}

func TestSMTPSenderReturnsCreateClientError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		_ = conn.Close()
	}()
	defer func() {
		_ = listener.Close()
		<-done
	}()

	host, port := splitSMTPAddr(t, listener.Addr().String())
	sender, err := NewSMTPSender(SMTPOptions{
		Host:       host,
		Port:       port,
		From:       "from@example.com",
		Encryption: EncryptionNone,
		Timeout:    time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	if err := sender.Send(context.Background(), validMessage()); err == nil {
		t.Fatal("expected create client error")
	}
}

func TestSMTPSenderReturnsSTARTTLSError(t *testing.T) {
	addr, _, stop := startFakeSMTPServer(t)
	defer stop()

	host, port := splitSMTPAddr(t, addr)
	sender, err := NewSMTPSender(SMTPOptions{
		Host:       host,
		Port:       port,
		From:       "from@example.com",
		Encryption: EncryptionSTARTTLS,
		Timeout:    time.Second,
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	if err := sender.Send(context.Background(), validMessage()); err == nil {
		t.Fatal("expected STARTTLS error")
	}
}

func TestSMTPSenderReturnsTLSError(t *testing.T) {
	addr, _, stop := startFakeSMTPServer(t)
	defer stop()

	host, port := splitSMTPAddr(t, addr)
	sender, err := NewSMTPSender(SMTPOptions{
		Host:       host,
		Port:       port,
		From:       "from@example.com",
		Encryption: EncryptionTLS,
		Timeout:    time.Second,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	if err != nil {
		t.Fatalf("NewSMTPSender returned error: %v", err)
	}

	if err := sender.Send(context.Background(), validMessage()); err == nil {
		t.Fatal("expected TLS error")
	}
}

func startFakeSMTPServer(t *testing.T) (string, <-chan smtpCapture, func()) {
	return startFakeSMTPServerWithReject(t, "")
}

func startFakeSMTPServerWithReject(t *testing.T, reject string) (string, <-chan smtpCapture, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen returned error: %v", err)
	}

	capture := make(chan smtpCapture, 1)
	done := make(chan struct{})
	var once sync.Once

	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			t.Errorf("Accept returned error: %v", err)
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)
		writeSMTPLine(t, writer, "220 localhost ESMTP")

		var recipients []string
		var data strings.Builder
		inData := false

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")

			if inData {
				if line == "." {
					if reject == "ENDDATA" {
						writeSMTPLine(t, writer, "550 data rejected")
						continue
					}
					writeSMTPLine(t, writer, "250 queued")
					capture <- smtpCapture{recipients: recipients, data: data.String()}
					inData = false
					continue
				}
				data.WriteString(line)
				data.WriteString("\r\n")
				continue
			}

			upper := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upper, "EHLO"):
				writeSMTPLine(t, writer, "250-localhost")
				if reject == "AUTH" {
					writeSMTPLine(t, writer, "250-AUTH PLAIN")
				}
				writeSMTPLine(t, writer, "250 OK")
			case strings.HasPrefix(upper, "AUTH"):
				if reject == "AUTH" {
					writeSMTPLine(t, writer, "535 auth failed")
					continue
				}
				writeSMTPLine(t, writer, "235 auth ok")
			case strings.HasPrefix(upper, "MAIL FROM:"):
				if reject == "MAIL" {
					writeSMTPLine(t, writer, "550 sender rejected")
					continue
				}
				writeSMTPLine(t, writer, "250 sender ok")
			case strings.HasPrefix(upper, "RCPT TO:"):
				if reject == "RCPT" {
					writeSMTPLine(t, writer, "550 recipient rejected")
					continue
				}
				recipients = append(recipients, extractSMTPPath(line))
				writeSMTPLine(t, writer, "250 recipient ok")
			case strings.HasPrefix(upper, "DATA"):
				if reject == "DATA" {
					writeSMTPLine(t, writer, "550 data rejected")
					continue
				}
				inData = true
				writeSMTPLine(t, writer, "354 end data with <CR><LF>.<CR><LF>")
			case strings.HasPrefix(upper, "QUIT"):
				if reject == "QUIT" {
					writeSMTPLine(t, writer, "550 quit rejected")
					return
				}
				writeSMTPLine(t, writer, "221 bye")
				return
			default:
				writeSMTPLine(t, writer, "250 ok")
			}
		}
	}()

	stop := func() {
		once.Do(func() {
			_ = listener.Close()
			<-done
		})
	}

	return listener.Addr().String(), capture, stop
}

func writeSMTPLine(t *testing.T, writer *bufio.Writer, line string) {
	t.Helper()

	if _, err := writer.WriteString(line + "\r\n"); err != nil {
		t.Logf("WriteString returned error: %v", err)
		return
	}
	if err := writer.Flush(); err != nil {
		t.Logf("Flush returned error: %v", err)
	}
}

func extractSMTPPath(line string) string {
	start := strings.Index(line, "<")
	end := strings.LastIndex(line, ">")
	if start == -1 || end == -1 || end <= start {
		return line
	}
	return line[start+1 : end]
}

func splitSMTPAddr(t *testing.T, addr string) (string, int) {
	t.Helper()

	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort returned error: %v", err)
	}
	var port int
	if _, err := fmt.Sscanf(portText, "%d", &port); err != nil {
		t.Fatalf("parse port returned error: %v", err)
	}
	return host, port
}
