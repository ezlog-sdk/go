package ezlog

import (
	"bytes"
	"compress/gzip"
)

func PointerConvert[T any](x T) *T {
	return &x
}

func GzipCompress(data *[]byte) (*[]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(*data); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return PointerConvert(buf.Bytes()), nil
}
