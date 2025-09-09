package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

// 压缩并编码
func CompressAndEncode(data []byte) (string, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return "", err
	}
	if err := zw.Close(); err != nil {
		return "", err
	}
	// Base64 编码
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// 解码并解压
func DecodeAndDecompress(encoded string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}
