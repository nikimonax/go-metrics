package httpsec

import (
	"bytes"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

type Hasher struct {
	header  string
	hashing func() hash.Hash
	key     []byte
}

func (hasher *Hasher) sign(content []byte) []byte {
	h := hmac.New(hasher.hashing, hasher.key)
	_, _ = h.Write(content)
	return h.Sum(nil)
}

func (hasher *Hasher) Sign(content []byte) string {
	return hex.EncodeToString(hasher.sign(content))
}

func (hasher *Hasher) Verify(content []byte, sig string) bool {
	macActual, err := hex.DecodeString(sig)

	if err != nil {
		return false
	}

	macExpected := hasher.sign(content)

	return hmac.Equal(macExpected, macActual)
}

func (hasher *Hasher) VerifyMiddleware(opts ...VerifyOption) httpextra.Middleware {
	var cfg VerifyConfig

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			hashHeader := r.Header.Get(hasher.header)
			hasHashHeader := hashHeader != ""

			if !hasHashHeader && !cfg.RequireHeader {
				next.ServeHTTP(w, r)
				return
			}

			content, err := httpextra.PeekContent(r)

			if err != nil {
				code := http.StatusInternalServerError
				http.Error(w, http.StatusText(code), code)
				return
			}

			if len(content) == 0 {
				if !hasHashHeader {
					next.ServeHTTP(w, r)
				} else {
					// if hash was provided without body -> error
					http.Error(w, ErrVerify.Error(), http.StatusBadRequest)
				}

				return
			}

			if !hasHashHeader {
				msg := fmt.Sprintf("required '%s' header", hasher.header)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}

			if !hasher.Verify(content, hashHeader) {
				http.Error(w, ErrVerify.Error(), http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(content))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (hasher *Hasher) CalculateHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := NewResponseWriterWithHashHeader(w, hasher.header, hasher.Sign)
		defer func() { _ = ww.Finalize() }()

		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}

func (hasher *Hasher) CalculateMiddleware() httpextra.Middleware {
	return hasher.CalculateHandler
}

func (hasher *Hasher) RoundTripper(next http.RoundTripper) http.RoundTripper {
	fn := func(req *http.Request) (*http.Response, error) {
		content, err := httpextra.PeekContent(req)

		if err != nil {
			return nil, err
		}

		if len(content) == 0 {
			return next.RoundTrip(req)
		}

		sig := hasher.Sign(content)

		newReq := req.Clone(req.Context())
		newReq.Header.Set(hasher.header, sig)

		return next.RoundTrip(newReq)
	}
	return httpextra.RoundTripperFunc(fn)
}

func NewHasher(
	header string,
	hashing func() hash.Hash,
	key []byte,
) *Hasher {
	return &Hasher{
		header:  header,
		hashing: hashing,
		key:     key,
	}
}
