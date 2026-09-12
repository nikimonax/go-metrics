package httpextra

import (
	"bytes"
	"io"
	"net/http"
)

func PeekContent(r *http.Request) ([]byte, error) {
	if r.Body == nil || r.Body == http.NoBody {
		return nil, nil
	}

	content, err := io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	if err := r.Body.Close(); err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(content))
	return content, nil
}
