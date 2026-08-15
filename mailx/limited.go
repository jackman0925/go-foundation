package mailx

import (
	"context"
	"errors"
)

// LimitOptions 配置并发限制发送器。
type LimitOptions struct {
	MaxConcurrent int
}

// LimitedSender 为已有 Sender 增加并发上限。
type LimitedSender struct {
	sender Sender
	slots  chan struct{}
}

// NewLimitedSender 创建并发受控的邮件发送器。
func NewLimitedSender(sender Sender, options LimitOptions) (*LimitedSender, error) {
	if sender == nil {
		return nil, errors.New("mail sender is required")
	}
	if options.MaxConcurrent <= 0 {
		return nil, errors.New("mail max concurrent must be greater than 0")
	}
	return &LimitedSender{
		sender: sender,
		slots:  make(chan struct{}, options.MaxConcurrent),
	}, nil
}

// Send 在可用并发槽位内发送邮件。
func (s *LimitedSender) Send(ctx context.Context, msg Message) error {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case s.slots <- struct{}{}:
		defer func() {
			<-s.slots
		}()
	case <-ctx.Done():
		return ctx.Err()
	}

	return s.sender.Send(ctx, msg)
}
