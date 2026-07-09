package compressx

import (
	"bytes"
	"compress/gzip"
)

// Gzip 使用 gzip 压缩数据。
func Gzip(input []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write(input); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Gunzip 解压 gzip 数据。
func Gunzip(input []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(reader); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
