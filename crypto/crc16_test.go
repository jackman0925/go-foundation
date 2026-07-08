package crypto

import "testing"

func TestCRC16Modbus(t *testing.T) {
	got := CRC16Modbus([]byte("123456789"))
	if got != 0x4B37 {
		t.Fatalf("expected 0x4B37, got 0x%04X", got)
	}
}
