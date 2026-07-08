package idgen

import (
	"testing"
	"time"
)

func TestNewSnowflakeRejectsInvalidMachineID(t *testing.T) {
	if _, err := NewSnowflake(SnowflakeOptions{MachineID: -1}); err == nil {
		t.Fatal("expected negative machine id error")
	}
	if _, err := NewSnowflake(SnowflakeOptions{MachineID: 1024}); err == nil {
		t.Fatal("expected too large machine id error")
	}
}

func TestSnowflakeNextIDIsMonotonic(t *testing.T) {
	generator, err := NewSnowflake(SnowflakeOptions{
		MachineID: 7,
		Epoch:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewSnowflake returned error: %v", err)
	}

	first, err := generator.NextID()
	if err != nil {
		t.Fatalf("NextID returned error: %v", err)
	}
	second, err := generator.NextID()
	if err != nil {
		t.Fatalf("NextID returned error: %v", err)
	}

	if second <= first {
		t.Fatalf("expected IDs to increase, got first=%d second=%d", first, second)
	}
}
