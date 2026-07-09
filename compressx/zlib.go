package compressx

import (
	"bytes"
	"compress/zlib"
)

// Zlib 使用 zlib 压缩数据。
func Zlib(input []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zlib.NewWriter(&buffer)
	if _, err := writer.Write(input); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Unzlib 解压 zlib 数据。
func Unzlib(input []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(input))
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
