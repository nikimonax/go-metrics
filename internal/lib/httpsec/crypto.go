package httpsec

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/ProtonMail/gopenpgp/v3/crypto"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

type Encryptor struct {
	pgp    *crypto.PGPHandle
	header string
}

type EncryptorOption = func(*Encryptor)

func NewEncryptor() *Encryptor {
	return &Encryptor{
		pgp:    crypto.PGP(),
		header: "X-Encrypted",
	}
}

func (e *Encryptor) DecryptMiddleware(key *crypto.Key) (httpextra.Middleware, error) {
	decHandle, err := e.pgp.Decryption().DecryptionKey(key).New()

	if err != nil {
		return nil, err
	}

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

			decrypted, err := decHandle.Decrypt(content, crypto.Bytes)

			if err != nil {
				http.Error(w, "failed decrypt payload", http.StatusBadRequest)
				return
			}

			newContent := decrypted.Bytes()

			r.Body = io.NopCloser(bytes.NewBuffer(newContent))
			r.ContentLength = int64(len(newContent))
			r.Header.Del(httpextra.HDRContentLength)
			r.Header.Del(e.header)

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}, nil
}

func (e *Encryptor) EncryptRoundTripper(
	next http.RoundTripper,
	key *crypto.Key,
) (http.RoundTripper, error) {
	encHandle, err := e.pgp.Encryption().Recipient(key).New()

	if err != nil {
		return nil, err
	}

	fn := func(req *http.Request) (resp *http.Response, err error) {
		content, err := httpextra.PeekContent(req)

		if err != nil {
			return nil, err
		}

		if len(content) == 0 {
			req.Body = nil
			return next.RoundTrip(req)
		}

		message, err := encHandle.Encrypt(content)

		if err != nil {
			return nil, err
		}

		newContent := message.Bytes()

		newReq := req.Clone(req.Context())
		newReq.Body = io.NopCloser(bytes.NewBuffer(newContent))
		newReq.ContentLength = int64(len(newContent))
		newReq.Header.Set(e.header, "true")

		return next.RoundTrip(newReq)
	}
	return httpextra.RoundTripperFunc(fn), nil
}
