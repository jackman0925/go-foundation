package mailx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"time"
)

const (
	defaultSMTPTimeout         = 10 * time.Second
	defaultSMTPMaxMessageBytes = 10 * 1024 * 1024
)

// Encryption 表示 SMTP 连接加密方式。
type Encryption int

const (
	// EncryptionNone 不启用连接加密。
	EncryptionNone Encryption = iota
	// EncryptionSTARTTLS 先建立明文连接，再升级为 TLS。
	EncryptionSTARTTLS
	// EncryptionTLS 建立隐式 TLS 连接，常用于 465 端口。
	EncryptionTLS
)

// SMTPOptions 配置 SMTP 邮件发送器。
type SMTPOptions struct {
	Host            string
	Port            int
	Username        string
	Password        string
	From            string
	Encryption      Encryption
	Timeout         time.Duration
	MaxMessageBytes int
	TLSConfig       *tls.Config
}

// SMTPSender 通过 SMTP 协议发送邮件。
type SMTPSender struct {
	options SMTPOptions
	dial    func(ctx context.Context) (*smtp.Client, error)
}

// NewSMTPSender 创建 SMTP 邮件发送器。
func NewSMTPSender(options SMTPOptions) (*SMTPSender, error) {
	if options.Host == "" {
		return nil, errors.New("mail smtp host is required")
	}
	if options.Port <= 0 {
		return nil, errors.New("mail smtp port must be greater than 0")
	}
	if options.From != "" {
		if _, err := mail.ParseAddress(options.From); err != nil {
			return nil, fmt.Errorf("mail smtp from is invalid: %w", err)
		}
	}
	if options.Timeout <= 0 {
		options.Timeout = defaultSMTPTimeout
	}
	if options.MaxMessageBytes == 0 {
		options.MaxMessageBytes = defaultSMTPMaxMessageBytes
	}
	if options.MaxMessageBytes < 0 {
		return nil, errors.New("mail smtp max message bytes must be greater than or equal to 0")
	}

	sender := &SMTPSender{options: options}
	sender.dial = sender.dialSMTP
	return sender, nil
}

// Send 发送邮件。
func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if msg.From == "" {
		msg.From = s.options.From
	}

	recipients, err := msg.Recipients()
	if err != nil {
		return err
	}
	data, err := buildMIMEMessage(msg)
	if err != nil {
		return err
	}
	if s.options.MaxMessageBytes > 0 && len(data) > s.options.MaxMessageBytes {
		return fmt.Errorf("mail message exceeds max size: %d > %d", len(data), s.options.MaxMessageBytes)
	}

	client, err := s.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	if s.options.Username != "" {
		auth := smtp.PlainAuth("", s.options.Username, s.options.Password, s.options.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("mail smtp auth failed: %w", err)
		}
	}

	from, err := mail.ParseAddress(msg.From)
	if err != nil {
		return fmt.Errorf("mail from is invalid: %w", err)
	}
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("mail smtp set sender failed: %w", err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("mail smtp set recipient failed: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("mail smtp open data failed: %w", err)
	}
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return fmt.Errorf("mail smtp write data failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("mail smtp close data failed: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("mail smtp quit failed: %w", err)
	}
	return nil
}

func (s *SMTPSender) dialSMTP(ctx context.Context) (*smtp.Client, error) {
	address := net.JoinHostPort(s.options.Host, fmt.Sprintf("%d", s.options.Port))
	timeoutCtx, cancel := context.WithTimeout(ctx, s.options.Timeout)
	defer cancel()

	var conn net.Conn
	var err error
	dialer := &net.Dialer{}

	switch s.options.Encryption {
	case EncryptionNone, EncryptionSTARTTLS:
		conn, err = dialer.DialContext(timeoutCtx, "tcp", address)
	case EncryptionTLS:
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config:    s.tlsConfig(),
		}
		conn, err = tlsDialer.DialContext(timeoutCtx, "tcp", address)
	default:
		return nil, errors.New("mail smtp encryption is invalid")
	}
	if err != nil {
		return nil, fmt.Errorf("mail smtp dial failed: %w", err)
	}

	if deadline, ok := timeoutCtx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	client, err := smtp.NewClient(conn, s.options.Host)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("mail smtp create client failed: %w", err)
	}

	if s.options.Encryption == EncryptionSTARTTLS {
		if err := client.StartTLS(s.tlsConfig()); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("mail smtp starttls failed: %w", err)
		}
	}

	return client, nil
}

func (s *SMTPSender) tlsConfig() *tls.Config {
	if s.options.TLSConfig != nil {
		return s.options.TLSConfig
	}
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: s.options.Host,
	}
}
