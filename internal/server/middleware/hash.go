package middleware

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

func VerifyHash(
	header string,
	hashing func() hash.Hash,
	key string,
) httpextra.Middleware {
	return NewHashVerifier(header, hashing, key).Handler
}

type HashVerifier struct {
	header  string
	hashing func() hash.Hash
	key     string
}

func (v *HashVerifier) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		content, err := io.ReadAll(r.Body)

		if err != nil {
			msg := fmt.Sprintf("failed read body: %s", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		// skip hash verify if there is no payload
		if len(content) == 0 {
			r.Body = io.NopCloser(bytes.NewReader(nil))
			next.ServeHTTP(w, r)
			return
		}

		hashExpected := r.Header.Get(v.header)
		if hashExpected == "" {
			msg := fmt.Sprintf("required '%s' header", v.header)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		h := v.hashing()

		_, err = io.Copy(h, io.MultiReader(
			bytes.NewReader([]byte(v.key)),
			bytes.NewReader(content),
		))

		if err != nil {
			http.Error(w, "failed calculate hash", http.StatusInternalServerError)
			return
		}

		hashActual := hex.EncodeToString(h.Sum(nil))

		if hashExpected != hashActual {
			http.Error(w, "invalid content hash", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(content))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func NewHashVerifier(
	header string,
	hashing func() hash.Hash,
	key string,
) *HashVerifier {
	return &HashVerifier{
		header:  header,
		hashing: hashing,
		key:     key,
	}
}

func CalculateHash(
	header string,
	hashing func() hash.Hash,
	key string,
) httpextra.Middleware {
	return NewHashCalculator(header, hashing, key).Handler
}

type HashCalculator struct {
	header  string
	hashing func() hash.Hash
	key     string
}

func (calc *HashCalculator) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		hash := calc.hashing()
		_, err := hash.Write([]byte(calc.key))

		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		ww := NewResponseWriterWithHashHeader(w, hash, calc.header)
		next.ServeHTTP(ww, r)
		_ = ww.finalize() // ignore errors
	}
	return http.HandlerFunc(fn)
}

func NewHashCalculator(
	header string,
	hashing func() hash.Hash,
	key string,
) *HashCalculator {
	return &HashCalculator{
		header:  header,
		hashing: hashing,
		key:     key,
	}
}

type ResponseWriterWithHashHeader struct {
	wrapped http.ResponseWriter
	header  string
	hash    hash.Hash
	body    bytes.Buffer
	code    int
}

// Header implements [http.ResponseWriter].
func (writer *ResponseWriterWithHashHeader) Header() http.Header {
	return writer.wrapped.Header()
}

// Write implements [http.ResponseWriter].
func (writer *ResponseWriterWithHashHeader) Write(data []byte) (int, error) {
	return io.MultiWriter(writer.hash, &writer.body).Write(data)
}

// WriteHeader implements [http.ResponseWriter].
func (writer *ResponseWriterWithHashHeader) WriteHeader(statusCode int) {
	writer.code = statusCode
}

func (writer *ResponseWriterWithHashHeader) finalize() error {
	hashString := hex.EncodeToString(writer.hash.Sum(nil))
	writer.wrapped.Header().Set(writer.header, hashString)
	writer.wrapped.WriteHeader(writer.code)
	_, err := writer.wrapped.Write(writer.body.Bytes())
	return err
}

func NewResponseWriterWithHashHeader(
	wrapped http.ResponseWriter,
	hash hash.Hash,
	header string,
) *ResponseWriterWithHashHeader {
	return &ResponseWriterWithHashHeader{
		wrapped: wrapped,
		hash:    hash,
		header:  header,
		code:    http.StatusOK,
	}
}
