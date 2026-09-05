package httpsec

import (
	"bytes"
	"net/http"
)

type ResponseWriterWithHashHeader struct {
	wrapped http.ResponseWriter
	header  string
	sign    func(content []byte) string
	body    bytes.Buffer
	code    int
}

// Header implements [http.ResponseWriter].
func (writer *ResponseWriterWithHashHeader) Header() http.Header {
	return writer.wrapped.Header()
}

// Write implements [http.ResponseWriter].
func (writer *ResponseWriterWithHashHeader) Write(data []byte) (int, error) {
	return writer.body.Write(data)
}

// WriteHeader implements [http.ResponseWriter].
func (writer *ResponseWriterWithHashHeader) WriteHeader(statusCode int) {
	writer.code = statusCode
}

func (writer *ResponseWriterWithHashHeader) Finalize() error {
	if writer.body.Len() > 0 {
		sig := writer.sign(writer.body.Bytes())
		writer.wrapped.Header().Set(writer.header, sig)
	}

	writer.wrapped.WriteHeader(writer.code)
	_, err := writer.wrapped.Write(writer.body.Bytes())
	return err
}

func NewResponseWriterWithHashHeader(
	wrapped http.ResponseWriter,
	header string,
	sign func(content []byte) string,
) *ResponseWriterWithHashHeader {
	return &ResponseWriterWithHashHeader{
		wrapped: wrapped,
		header:  header,
		sign:    sign,
		code:    http.StatusOK,
	}
}
