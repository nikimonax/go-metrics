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

		if _, err := h.Write(content); err != nil {
			http.Error(w, "failed calculate hash", http.StatusInternalServerError)
			return
		}

		if _, err := h.Write([]byte(v.key)); err != nil {
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
