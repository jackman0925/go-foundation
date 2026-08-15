package mailx

import (
	"context"
	"errors"
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
	if _, err := NewLimitedSender(base, LimitOptions{MaxConcurrent: 1, MaxWaiting: -1}); err == nil {
		t.Fatal("expected invalid max waiting error")
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

func TestLimitedSenderQueuesWhenWaitingCapacityAvailable(t *testing.T) {
	base := &blockingSender{
		started: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	sender, err := NewLimitedSender(base, LimitOptions{MaxConcurrent: 1, MaxWaiting: 1})
	if err != nil {
		t.Fatalf("NewLimitedSender returned error: %v", err)
	}

	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)

	go func() {
		firstDone <- sender.Send(context.Background(), validMessage())
	}()
	<-base.started

	go func() {
		secondDone <- sender.Send(context.Background(), validMessage())
	}()
	waitUntil(t, func() bool {
		return len(sender.waiting) == 1
	})

	select {
	case <-base.started:
		t.Fatal("queued send started before concurrency slot was released")
	case <-time.After(30 * time.Millisecond):
	}

	close(base.release)

	if err := <-firstDone; err != nil {
		t.Fatalf("first Send returned error: %v", err)
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second Send returned error: %v", err)
	}
	if got := atomic.LoadInt32(&base.maxSeen); got > 1 {
		t.Fatalf("expected max concurrency <= 1, got %d", got)
	}
}

func TestLimitedSenderReturnsBusyWhenWaitingQueueFull(t *testing.T) {
	base := &blockingSender{
		started: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	sender, err := NewLimitedSender(base, LimitOptions{MaxConcurrent: 1, MaxWaiting: 1})
	if err != nil {
		t.Fatalf("NewLimitedSender returned error: %v", err)
	}

	firstDone := make(chan error, 1)
	secondDone := make(chan error, 1)

	go func() {
		firstDone <- sender.Send(context.Background(), validMessage())
	}()
	<-base.started

	go func() {
		secondDone <- sender.Send(context.Background(), validMessage())
	}()
	waitUntil(t, func() bool {
		return len(sender.waiting) == 1
	})

	err = sender.Send(context.Background(), validMessage())
	if !errors.Is(err, ErrMailBusy) {
		t.Fatalf("expected ErrMailBusy, got %v", err)
	}

	close(base.release)

	if err := <-firstDone; err != nil {
		t.Fatalf("first Send returned error: %v", err)
	}
	if err := <-secondDone; err != nil {
		t.Fatalf("second Send returned error: %v", err)
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

func TestNewSMTPClientUsesSafeDefaults(t *testing.T) {
	sender, err := NewSMTPClient(SMTPClientOptions{
		SMTP: SMTPOptions{
			Host: "127.0.0.1",
			Port: 25,
			From: "from@example.com",
		},
	})
	if err != nil {
		t.Fatalf("NewSMTPClient returned error: %v", err)
	}

	limited, ok := sender.(*LimitedSender)
	if !ok {
		t.Fatalf("expected *LimitedSender, got %T", sender)
	}
	if cap(limited.slots) != defaultSMTPClientMaxConcurrent {
		t.Fatalf("expected default max concurrent %d, got %d", defaultSMTPClientMaxConcurrent, cap(limited.slots))
	}
	if cap(limited.waiting) != defaultSMTPClientMaxWaiting {
		t.Fatalf("expected default max waiting %d, got %d", defaultSMTPClientMaxWaiting, cap(limited.waiting))
	}
}

func TestNewSMTPClientUsesCustomLimits(t *testing.T) {
	sender, err := NewSMTPClient(SMTPClientOptions{
		SMTP: SMTPOptions{
			Host: "127.0.0.1",
			Port: 25,
			From: "from@example.com",
		},
		LimitOptions: LimitOptions{
			MaxConcurrent: 2,
			MaxWaiting:    3,
		},
	})
	if err != nil {
		t.Fatalf("NewSMTPClient returned error: %v", err)
	}

	limited := sender.(*LimitedSender)
	if cap(limited.slots) != 2 {
		t.Fatalf("expected max concurrent 2, got %d", cap(limited.slots))
	}
	if cap(limited.waiting) != 3 {
		t.Fatalf("expected max waiting 3, got %d", cap(limited.waiting))
	}
}

func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not met before deadline")
}

func validMessage() Message {
	return Message{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "hello",
		Text:    "content",
	}
}
