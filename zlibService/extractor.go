package zlibService

import (
	"bytes"
	"compress/zlib"
	"io"

	"github.com/gookit/slog"
)

type ZlibExtractor struct{}

func (z ZlibExtractor) Extract(src []byte) ([]byte, error) {
	byteBuff := bytes.NewReader(src)
	reader, err := zlib.NewReader(byteBuff)
	if err != nil {
		slog.Errorf("zlib.NewReader failed: %w", err)
		return nil, err
	}
	defer reader.Close()
	outBuff := bytes.NewBuffer([]byte{})
	io.Copy(outBuff, reader)
	return outBuff.Bytes(), nil
}

func NewZlibExtractor() *ZlibExtractor {
	return &ZlibExtractor{}
}
