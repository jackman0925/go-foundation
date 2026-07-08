package crypto

// CRC16Modbus returns the CRC-16/MODBUS checksum.
func CRC16Modbus(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, item := range data {
		crc ^= uint16(item)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}
