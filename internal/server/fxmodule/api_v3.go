package fxmodule

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/lib/httpsec"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func APIV3Module() fx.Option {
	return fx.Module(
		"api_v3",
		fx.Provide(
			fx.Annotate(
				handler.NewUpdateMetricsV3Handler,
				fx.ParamTags("", `name:"presenter_error_json"`, ""),
				fx.ResultTags(`name:"handler_metrics_update_v3"`),
				fx.As(new(http.Handler)),
			),
			fx.Annotate(
				provideJSONRouter,
				fx.ParamTags(
					`name:"router_base"`,
					`name:"middleware_logger"`,
					`name:"middleware_compress"`,
				),
				fx.ResultTags(`name:"router_v3"`),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerHandlersV3,
				fx.ParamTags(
					`name:"router_v3"`,
					`name:"handler_metrics_update_v3"`,
					"",
				),
			),
		),
	)
}

func registerHandlersV3(
	router chi.Router,
	updateMetricsHandler http.Handler,
	cfg *config.ServerConfig,
	hasher *httpsec.Hasher,
	encryptor *httpsec.Encryptor,
) error {
	if hasher != nil {
		verify := hasher.VerifyMiddleware(httpsec.RequireHeader())
		router = router.With(verify)
	}

	if cfg.Security.CryptoKeyFile != "" {
		key, err := cfg.Security.ReadPrivateKey()

		if err != nil {
			return err
		}

		decrypt, err := encryptor.DecryptMiddleware(key)

		if err != nil {
			return err
		}

		router = router.With(decrypt)
	}

	router.Post(
		"/updates",
		updateMetricsHandler.ServeHTTP,
	)

	return nil
}
