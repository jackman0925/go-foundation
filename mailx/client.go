package mailx

const (
	defaultSMTPClientMaxConcurrent = 5
	defaultSMTPClientMaxWaiting    = 100
)

// SMTPClientOptions 配置默认带限流保护的 SMTP 客户端。
type SMTPClientOptions struct {
	SMTP SMTPOptions
	LimitOptions
}

// NewSMTPClient 创建默认带并发和等待队列保护的 SMTP 发送器。
func NewSMTPClient(options SMTPClientOptions) (Sender, error) {
	if options.MaxConcurrent == 0 {
		options.MaxConcurrent = defaultSMTPClientMaxConcurrent
	}
	if options.MaxWaiting == 0 {
		options.MaxWaiting = defaultSMTPClientMaxWaiting
	}

	smtpSender, err := NewSMTPSender(options.SMTP)
	if err != nil {
		return nil, err
	}
	return NewLimitedSender(smtpSender, options.LimitOptions)
}
