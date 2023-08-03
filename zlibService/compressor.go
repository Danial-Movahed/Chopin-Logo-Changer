package zlibService

import (
	"bytes"
	"compress/zlib"
	"io"
)

type Compressor struct{}

func (c Compressor) Compress(data []byte) ([]byte, error) {
	var out bytes.Buffer
	writer, err := zlib.NewWriterLevel(&out, 3)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	dataReader := bytes.NewReader(data)
	io.Copy(writer, dataReader)
	return out.Bytes(), nil
}

func NewZlibCompressor() *Compressor {
	return &Compressor{}
}
