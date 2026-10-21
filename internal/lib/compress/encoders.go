package compress

import (
	"errors"
	"io"

	"github.com/klauspost/compress/flate"
	"github.com/klauspost/compress/gzip"
)

func EncoderGzip(w io.Writer, level int) io.Writer {
	gw, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return nil
	}
	return gw
}

func EncoderDeflate(w io.Writer, level int) io.Writer {
	dw, err := flate.NewWriter(w, level)
	if err != nil {
		return nil
	}
	return dw
}

func ResetEncoder(encoder io.Writer, w io.Writer) error {
	resetter, ok := encoder.(interface{ Reset(io.Writer) })

	if !ok {
		return errors.New("encoder does not support reset")
	}

	resetter.Reset(w)

	return nil
}
