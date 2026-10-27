package compress

import (
	"io"

	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
)

func DecoderGzip(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func DecoderDeflate(r io.Reader) (io.ReadCloser, error) {
	return flate.NewReader(r), nil
}
