package httpsec

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

type Encryptor struct {
	hashing func() hash.Hash
	random  io.Reader
	header  string
}

type EncryptorOption = func(*Encryptor)

func WithHashing(hashing func() hash.Hash) EncryptorOption {
	return func(e *Encryptor) {
		e.hashing = hashing
	}
}

func WithRandom(random io.Reader) EncryptorOption {
	return func(e *Encryptor) {
		e.random = random
	}
}

func NewEncryptor(options ...EncryptorOption) *Encryptor {
	e := &Encryptor{
		hashing: sha512.New,
		random:  rand.Reader,
		header:  "X-Encrypted",
	}

	for _, opt := range options {
		opt(e)
	}

	return e
}

func (e *Encryptor) encrypt(content []byte, key *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(
		e.hashing(),
		e.random,
		key,
		content,
		nil,
	)
}

func (e *Encryptor) decrypt(content []byte, key *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptOAEP(
		e.hashing(),
		e.random,
		key,
		content,
		nil,
	)
}

func (e *Encryptor) DecryptMiddleware(key *rsa.PrivateKey) httpextra.Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			isEncryptedRaw := r.Header.Get(e.header)

			if isEncryptedRaw == "" {
				next.ServeHTTP(w, r)
				return
			}

			isEncrypted, err := strconv.ParseBool(isEncryptedRaw)

			if err != nil {
				msg := fmt.Sprintf("Bad '%s' header: '%s'", e.header, isEncryptedRaw)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}

			if !isEncrypted {
				next.ServeHTTP(w, r)
				return
			}

			content, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, "failed read body", http.StatusInternalServerError)
				return
			}

			decrypted, err := e.decrypt(content, key)

			if err != nil {
				http.Error(w, "failed decrypt payload", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(decrypted))
			r.ContentLength = int64(len(decrypted))
			r.Header.Del(httpextra.HDRContentLength)
			r.Header.Del(e.header)

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (e *Encryptor) EncryptRoundTripper(next http.RoundTripper, key *rsa.PublicKey) http.RoundTripper {
	fn := func(req *http.Request) (*http.Response, error) {
		content, err := httpextra.PeekContent(req)

		if err != nil {
			return nil, err
		}

		if len(content) == 0 {
			req.Body = nil
			return next.RoundTrip(req)
		}

		newContent, err := e.encrypt(content, key)

		if err != nil {
			return nil, err
		}

		newReq := req.Clone(req.Context())
		newReq.Body = io.NopCloser(bytes.NewBuffer(newContent))
		newReq.ContentLength = int64(len(newContent))
		newReq.Header.Set(e.header, "true")

		return next.RoundTrip(newReq)
	}
	return httpextra.RoundTripperFunc(fn)
}
