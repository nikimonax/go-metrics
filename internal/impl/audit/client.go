package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sethvargo/go-retry"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
)

type Client interface {
	Send(ctx context.Context, record Record) error
}

type FileClient struct {
	mu      sync.Mutex
	file    *os.File
	encoder *json.Encoder
}

// Send implements [Client].
func (client *FileClient) Send(_ context.Context, record Record) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.file == nil {
		return errors.New("file closed")
	}

	return client.encoder.Encode(record)
}

func (client *FileClient) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if err := client.file.Close(); err != nil {
		return err
	}

	client.file = nil
	client.encoder = nil
	return nil
}

var _ Client = (*FileClient)(nil)

func NewFileClient(filePath string) (*FileClient, error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		return nil, err
	}

	encoder := json.NewEncoder(file)

	return &FileClient{
		file:    file,
		encoder: encoder,
	}, nil
}

const defaultRequestTimeout = time.Second

type HTTPClient struct {
	http     *http.Client
	endpoint string
	timeout  time.Duration
}

// Send implements [Client].
func (client *HTTPClient) Send(ctx context.Context, record Record) (err error) {
	content := new(bytes.Buffer)

	if err = json.NewEncoder(content).Encode(record); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.endpoint, content)

	if err != nil {
		return fmt.Errorf("failed create request: %w", err)
	}

	req.Header.Set(httpextra.HDRContentType, httpextra.MIMEJSON)

	resp, err := client.http.Do(req)

	if err != nil {
		return fmt.Errorf("failed send audit record: network error: %w", err)
	}

	defer func() {
		if _, deferErr := io.Copy(io.Discard, resp.Body); deferErr != nil {
			err = errors.Join(err, deferErr)
		}

		if deferErr := resp.Body.Close(); deferErr != nil {
			err = errors.Join(err, deferErr)
		}
	}()

	if resp.StatusCode >= 400 {
		reason := "unknown"

		if body, err := io.ReadAll(resp.Body); err == nil && len(body) > 0 {
			reason = string(body)
		}

		return httpextra.NewErrHTTPStatus(resp.StatusCode, reason)
	}

	return nil
}

var _ Client = (*HTTPClient)(nil)

func NewHTTPClient(endpoint string) *HTTPClient {
	return &HTTPClient{
		http:     &http.Client{},
		endpoint: endpoint,
		timeout:  defaultRequestTimeout,
	}
}

type RetryClient struct {
	wrapped     Client
	backoff     func() retry.Backoff
	isRetryable func(error) bool
}

// Send implements [Client].
func (client *RetryClient) Send(ctx context.Context, record Record) error {
	return retry.Do(ctx, client.backoff(), func(ctx context.Context) error {
		err := client.wrapped.Send(ctx, record)

		if err != nil && client.isRetryable(err) {
			err = retry.RetryableError(err)
		}

		return err
	})
}

var _ Client = (*RetryClient)(nil)

func NewRetryClient(
	client Client,
	backoff func() retry.Backoff,
	isRetryable func(error) bool,
) *RetryClient {
	return &RetryClient{
		wrapped:     client,
		backoff:     backoff,
		isRetryable: isRetryable,
	}
}

type ClientGroup struct {
	clients []Client
}

// Send implements [Client].
func (group *ClientGroup) Send(ctx context.Context, record Record) error {
	errChan := make(chan error, len(group.clients))

	for _, client := range group.clients {
		go func() {
			errChan <- client.Send(ctx, record)
		}()
	}

	var errs error
	for range len(group.clients) {
		errs = errors.Join(errs, <-errChan)
	}

	return errs
}

var _ Client = (*ClientGroup)(nil)

func NewClientGroup(clients []Client) *ClientGroup {
	return &ClientGroup{clients: clients}
}
