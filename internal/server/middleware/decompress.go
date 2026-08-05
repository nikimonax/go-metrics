package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/lib/ioextra"
)

func Decompress() func(http.Handler) http.Handler {
	return NewDecompressor().Handler
}

type DecoderFunc func(io.Reader) (io.ReadCloser, error)

type Decompressor struct {
	decoders map[string]DecoderFunc
}

func NewDecompressor() *Decompressor {
	d := &Decompressor{
		decoders: make(map[string]DecoderFunc),
	}

	d.SetDecoder("gzip", decoderGzip)

	return d
}

func (d *Decompressor) SetDecoder(encoding string, fn DecoderFunc) {
	encoding = normalizeEncoding(encoding)

	if encoding == "" {
		panic("the encoding can not be empty")
	}

	if fn == nil {
		panic("attempted to set a nil decoder function")
	}

	d.decoders[encoding] = fn
}

func decoderGzip(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func (d *Decompressor) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		encoding := normalizeEncoding(r.Header.Get(httpextra.HDRContentEncoding))
		r.Header.Del(httpextra.HDRContentEncoding)

		// если тело не закодировано - пропускаем
		if encoding == "" || encoding == "identity" {
			next.ServeHTTP(w, r)
			return
		}

		// если тела запроса нет - пропускаем
		if r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}

		decoder, ok := d.decoders[encoding]

		if !ok {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		newBody, err := decoder(r.Body)

		if err != nil {
			msg := fmt.Sprintf("failed decode payload using '%s' decoder", encoding)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		// Если следующий handler вызывает r.Body.Close(),
		// то закрываются сначала декодер, потом body
		// либо в этом скоупе только декодер с помощью defer

		// Идемпотентный closer нужен, потому что закрытие декодера
		// может произойти два раза:
		// 	1 - defer;
		// 	2 - потенциально в следующем хендлере r.Body.Close().

		decoderCloser := ioextra.NewIdempotentCloser(newBody)
		defer func() {
			if err := decoderCloser.Close(); err != nil {
				// TODO: заменить на zap sugar
				log.Printf("failed close decoder: %s", err)
			}
		}()

		r.Body = ioextra.NewReadCloser(newBody, ioextra.NewMultiCloser(decoderCloser, r.Body))
		r.Header.Del(httpextra.HDRContentLength)
		r.ContentLength = -1

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func normalizeEncoding(encoding string) string {
	encoding = strings.TrimSpace(encoding)
	encoding = strings.ToLower(encoding)
	return encoding
}
