package mailx

import (
	"context"
	"errors"
)

const defaultMaxWaiting = 100

// ErrMailBusy 表示发送器已达到并发和等待队列上限，邮件没有发送也没有入队。
var ErrMailBusy = errors.New("mail sender is busy")

// LimitOptions 配置并发限制发送器。
type LimitOptions struct {
	MaxConcurrent int
	MaxWaiting    int
}

// LimitedSender 为已有 Sender 增加并发和等待队列上限。
type LimitedSender struct {
	sender  Sender
	slots   chan struct{}
	waiting chan struct{}
}

// NewLimitedSender 创建并发受控的邮件发送器。
func NewLimitedSender(sender Sender, options LimitOptions) (*LimitedSender, error) {
	if sender == nil {
		return nil, errors.New("mail sender is required")
	}
	if options.MaxConcurrent <= 0 {
		return nil, errors.New("mail max concurrent must be greater than 0")
	}
	if options.MaxWaiting < 0 {
		return nil, errors.New("mail max waiting must be greater than or equal to 0")
	}
	if options.MaxWaiting == 0 {
		options.MaxWaiting = defaultMaxWaiting
	}
	return &LimitedSender{
		sender:  sender,
		slots:   make(chan struct{}, options.MaxConcurrent),
		waiting: make(chan struct{}, options.MaxWaiting),
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
		return s.sender.Send(ctx, msg)
	default:
	}

	select {
	case s.waiting <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrMailBusy
	}

	select {
	case s.slots <- struct{}{}:
		<-s.waiting
		defer func() {
			<-s.slots
		}()
	case <-ctx.Done():
		<-s.waiting
		return ctx.Err()
	}

	return s.sender.Send(ctx, msg)
}
