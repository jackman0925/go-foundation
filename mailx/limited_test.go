package mailx

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type blockingSender struct {
	started chan struct{}
	release chan struct{}
	active  int32
	maxSeen int32
}

func (s *blockingSender) Send(ctx context.Context, msg Message) error {
	current := atomic.AddInt32(&s.active, 1)
	defer atomic.AddInt32(&s.active, -1)

	for {
		maxSeen := atomic.LoadInt32(&s.maxSeen)
		if current <= maxSeen || atomic.CompareAndSwapInt32(&s.maxSeen, maxSeen, current) {
			break
		}
	}

	s.started <- struct{}{}

	select {
	case <-s.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestNewLimitedSenderRejectsInvalidOptions(t *testing.T) {
	if _, err := NewLimitedSender(nil, LimitOptions{MaxConcurrent: 1}); err == nil {
		t.Fatal("expected nil sender error")
	}

	base := &blockingSender{}
	if _, err := NewLimitedSender(base, LimitOptions{MaxConcurrent: 0}); err == nil {
		t.Fatal("expected invalid max concurrent error")
	}
}

func TestLimitedSenderLimitsConcurrency(t *testing.T) {
	base := &blockingSender{
		started: make(chan struct{}, 3),
		release: make(chan struct{}),
	}
	sender, err := NewLimitedSender(base, LimitOptions{MaxConcurrent: 2})
	if err != nil {
		t.Fatalf("NewLimitedSender returned error: %v", err)
	}

	ctx := context.Background()
	done := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func() {
			done <- sender.Send(ctx, validMessage())
		}()
	}

	<-base.started
	<-base.started

	select {
	case <-base.started:
		t.Fatal("third send started before concurrency slot was released")
	case <-time.After(30 * time.Millisecond):
	}

	close(base.release)

	for i := 0; i < 3; i++ {
		if err := <-done; err != nil {
			t.Fatalf("Send returned error: %v", err)
		}
	}
	if got := atomic.LoadInt32(&base.maxSeen); got > 2 {
		t.Fatalf("expected max concurrency <= 2, got %d", got)
	}
}

func TestLimitedSenderReturnsContextErrorWhenWaitingForSlot(t *testing.T) {
	base := &blockingSender{
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	sender, err := NewLimitedSender(base, LimitOptions{MaxConcurrent: 1})
	if err != nil {
		t.Fatalf("NewLimitedSender returned error: %v", err)
	}

	firstDone := make(chan error, 1)
	go func() {
		firstDone <- sender.Send(context.Background(), validMessage())
	}()
	<-base.started

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := sender.Send(ctx, validMessage()); err == nil {
		t.Fatal("expected context error")
	}

	close(base.release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first Send returned error: %v", err)
	}
}

func TestLimitedSenderAllowsNilContext(t *testing.T) {
	base := &blockingSender{
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	sender, err := NewLimitedSender(base, LimitOptions{MaxConcurrent: 1})
	if err != nil {
		t.Fatalf("NewLimitedSender returned error: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- sender.Send(nil, validMessage())
	}()
	<-base.started
	close(base.release)

	if err := <-done; err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
}

func validMessage() Message {
	return Message{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "hello",
		Text:    "content",
	}
}
