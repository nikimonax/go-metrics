package shared

import (
	"context"
	"log"
	"net/http/httptest"
	"net/url"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap/zapcore"

	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/fxmodule"
)

type TB interface {
	Helper()
	Fatal(args ...any)
}

type nopTB struct{}

func (tb *nopTB) Helper() {}

func (tb *nopTB) Fatal(args ...any) {
	log.Fatalln(args...)
}

var _ TB = (*nopTB)(nil)

func NewTestServer(tb TB) (*url.URL, func()) {
	if tb == nil {
		tb = new(nopTB)
	}

	tb.Helper()

	cfg := config.NewDefaultConfig()
	cfg.Listen = "127.0.0.1:0"
	cfg.Profile = true
	cfg.Log.Level = zapcore.ErrorLevel

	var router chi.Router

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.Supply(&cfg),
		fxmodule.StorageModule(&cfg),
		fxmodule.CoreModule(),
		fxmodule.ProfileModule(&cfg),
		fxmodule.APIV1Module(),
		fxmodule.APIV2Module(),
		fxmodule.APIV3Module(),
		fx.Populate(
			fx.Annotate(
				&router,
				fx.ParamTags(`name:"router_base"`),
			),
		),
	)

	if err := app.Err(); err != nil {
		tb.Fatal(err)
	}

	if err := app.Start(context.Background()); err != nil {
		tb.Fatal(err)
	}

	server := httptest.NewServer(router)
	baseURL, err := url.Parse(server.URL)

	if err != nil {
		tb.Fatal(err)
	}

	return baseURL, func() {
		server.Close()
		_ = app.Stop(context.Background())
	}
}
