package idgen

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewSafeNumberIDRejectsInvalidSequenceDigits(t *testing.T) {
	if _, err := NewSafeNumberID(SafeNumberIDOptions{SequenceDigits: -1}); err == nil {
		t.Fatal("expected negative sequence digits error")
	}
	if _, err := NewSafeNumberID(SafeNumberIDOptions{SequenceDigits: 5}); err == nil {
		t.Fatal("expected too large sequence digits error")
	}
}

func TestNewSafeNumberIDUsesDefaultSequenceDigits(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{Location: time.UTC})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	generator.now = func() time.Time {
		return time.Date(2026, 8, 12, 9, 8, 7, 0, time.UTC)
	}

	id, err := generator.NextString()
	if err != nil {
		t.Fatalf("NextString returned error: %v", err)
	}
	if id != "26224090807000" {
		t.Fatalf("expected default sequence digits id, got %s", id)
	}
}

func TestNewSafeNumberIDUsesDefaultLocation(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	if generator.location != time.Local {
		t.Fatal("expected default location to use time.Local")
	}
}

func TestSafeNumberIDNextStringUsesExpectedFormat(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{
		SequenceDigits: 3,
		Location:       time.UTC,
	})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	generator.now = func() time.Time {
		return time.Date(2026, 8, 12, 9, 8, 7, 0, time.UTC)
	}

	id, err := generator.NextString()
	if err != nil {
		t.Fatalf("NextString returned error: %v", err)
	}
	if id != "26224090807000" {
		t.Fatalf("expected formatted id 26224090807000, got %s", id)
	}
}

func TestSafeNumberIDNextIsJSSafeInteger(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{SequenceDigits: 4, Location: time.UTC})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	generator.now = func() time.Time {
		return time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	id, err := generator.Next()
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if id > maxJSSafeInteger {
		t.Fatalf("expected JS safe integer, got %d", id)
	}
}

func TestSafeNumberIDIncrementsSequenceWithinSameSecond(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{SequenceDigits: 2, Location: time.UTC})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	generator.now = func() time.Time {
		return time.Date(2026, 8, 12, 9, 8, 7, 0, time.UTC)
	}

	first, err := generator.NextString()
	if err != nil {
		t.Fatalf("first NextString returned error: %v", err)
	}
	second, err := generator.NextString()
	if err != nil {
		t.Fatalf("second NextString returned error: %v", err)
	}

	if first != "2622409080700" {
		t.Fatalf("expected first id sequence 00, got %s", first)
	}
	if second != "2622409080701" {
		t.Fatalf("expected second id sequence 01, got %s", second)
	}
}

func TestSafeNumberIDResetsSequenceOnNewSecond(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{SequenceDigits: 2, Location: time.UTC})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	times := []time.Time{
		time.Date(2026, 8, 12, 9, 8, 7, 0, time.UTC),
		time.Date(2026, 8, 12, 9, 8, 8, 0, time.UTC),
	}
	generator.now = func() time.Time {
		next := times[0]
		times = times[1:]
		return next
	}

	if _, err := generator.NextString(); err != nil {
		t.Fatalf("first NextString returned error: %v", err)
	}
	second, err := generator.NextString()
	if err != nil {
		t.Fatalf("second NextString returned error: %v", err)
	}

	if second != "2622409080800" {
		t.Fatalf("expected sequence reset on new second, got %s", second)
	}
}

func TestSafeNumberIDReturnsErrorWhenSequenceOverflows(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{SequenceDigits: 1, Location: time.UTC})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	generator.now = func() time.Time {
		return time.Date(2026, 8, 12, 9, 8, 7, 0, time.UTC)
	}

	for i := 0; i < 10; i++ {
		if _, err := generator.NextString(); err != nil {
			t.Fatalf("NextString %d returned error: %v", i, err)
		}
	}
	if _, err := generator.NextString(); err == nil {
		t.Fatal("expected sequence overflow error")
	}
}

func TestSafeNumberIDNextReturnsErrorWhenSequenceOverflows(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{SequenceDigits: 1, Location: time.UTC})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	generator.now = func() time.Time {
		return time.Date(2026, 8, 12, 9, 8, 7, 0, time.UTC)
	}

	for i := 0; i < 10; i++ {
		if _, err := generator.Next(); err != nil {
			t.Fatalf("Next %d returned error: %v", i, err)
		}
	}
	if _, err := generator.Next(); err == nil {
		t.Fatal("expected sequence overflow error")
	}
}

func TestSafeNumberIDConcurrentCallsAreUnique(t *testing.T) {
	generator, err := NewSafeNumberID(SafeNumberIDOptions{SequenceDigits: 4, Location: time.UTC})
	if err != nil {
		t.Fatalf("NewSafeNumberID returned error: %v", err)
	}
	generator.now = func() time.Time {
		return time.Date(2026, 8, 12, 9, 8, 7, 0, time.UTC)
	}

	const count = 200
	var wg sync.WaitGroup
	ids := make(chan string, count)
	errs := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := generator.NextString()
			if err != nil {
				errs <- err
				return
			}
			ids <- id
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		t.Fatalf("NextString returned error: %v", err)
	}

	seen := make(map[string]struct{}, count)
	for id := range ids {
		if _, err := strconv.ParseInt(id, 10, 64); err != nil {
			t.Fatalf("id is not numeric: %s", id)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicated id: %s", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != count {
		t.Fatalf("expected %d ids, got %d", count, len(seen))
	}
}
