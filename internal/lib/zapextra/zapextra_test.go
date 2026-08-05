package zapextra_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikimonax/go-metrics/internal/lib/zapextra"
	"github.com/nikimonax/go-metrics/internal/testing/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapLoggerFactory(t *testing.T) {
	type TestCase struct {
		name  string
		env   zapextra.LogEnv
		panic bool
	}

	tests := []TestCase{
		{
			name:  "development",
			env:   zapextra.EnvDev,
			panic: false,
		},
		{
			name:  "production",
			env:   zapextra.EnvProd,
			panic: false,
		},
		{
			name:  "unknown",
			env:   zapextra.LogEnv(-1),
			panic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.panic {
				assert.PanicsWithError(t, zapextra.ErrUnknownEnv.Error(), func() {
					zapextra.NewZapLogger(tc.env)
				})
			} else {
				logger := zapextra.NewZapLogger(tc.env)
				require.NotNil(t, logger)
			}
		})
	}
}

func TestZapLoggerMiddleware(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	defer func() { assert.NoError(t, logger.Sync()) }()

	middleware := zapextra.NewZapSugarLoggingMiddleware(logger)
	handlerFunc := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("body"))
	}

	handler := middleware(http.HandlerFunc(handlerFunc))

	req := shared.NewRequest(t, http.MethodPost, nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "handle request", entry.Message)
	assert.Equal(t, "POST", entry.ContextMap()["method"])
	assert.Equal(t, shared.DefaultURL, entry.ContextMap()["uri"])
	assert.Equal(t, int64(http.StatusCreated), entry.ContextMap()["status"])
	assert.Equal(t, int64(len("body")), entry.ContextMap()["size"])
}
